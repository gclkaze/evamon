package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/gclkaze/evamon/cmd/internal/models"
	port "github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/pkg/utils"
)

type MultiSeriesRing struct {
	mu sync.RWMutex

	maxPoints int
	vars      int

	times  []time.Time
	values [][]int

	owner    port.IDiagram
	varnames []string
}

func NewMultiSeriesRing(maxPoints int, vars int, owner port.IDiagram, varnames []string) *MultiSeriesRing {
	if maxPoints <= 1 {
		maxPoints = 50
	}
	if vars <= 0 {
		vars = 1
	}

	ms := &MultiSeriesRing{
		maxPoints: maxPoints,
		vars:      vars,
		times:     make([]time.Time, 0, maxPoints),
		values:    make([][]int, vars),
		owner:     owner,
		varnames:  append([]string{}, varnames...),
	}

	for i := 0; i < vars; i++ {
		ms.values[i] = make([]int, 0, maxPoints)
	}

	ms.initializeFilter()

	return ms
}

func (m *MultiSeriesRing) initializeFilter() {
	theFilter := m.owner.GetFilter()
	if theFilter != nil {
		setup := theFilter.Setup
		if setup != nil {
			// just pre-allocate capacity, no pre-filling
			setup.ConditionResults = make([]models.ConditionResult, 0, m.maxPoints)
		}
	}
}

func (m *MultiSeriesRing) Vars() int {
	return m.vars
}

func (m *MultiSeriesRing) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.times)
}

func (m *MultiSeriesRing) MaxPoints() int {
	return m.maxPoints
}

func (m *MultiSeriesRing) Append(at time.Time, nums []int) {
	if len(nums) != m.vars {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.times = append(m.times, at)

	for i := 0; i < m.vars; i++ {
		v := nums[i]
		if v < 0 {
			v = 0
		}
		m.values[i] = append(m.values[i], v)
	}

	floats := make([]float64, len(nums))
	for i, v := range nums {
		floats[i] = float64(v)
	}

	last := len(m.times) - 1
	theFilter := m.owner.GetFilter()
	if theFilter != nil {
		setup := theFilter.Setup
		if setup != nil && len(setup.Components) > 0 {
			// grow ConditionResults in sync with times
			setup.ConditionResults = append(setup.ConditionResults, models.ConditionResult{
				Results: make([]bool, len(setup.Components)),
			})

			for i := range setup.Components {
				if !setup.Components[i].Enabled {
					continue
				}
				res, err := utils.RunExpression(setup.Components[i].Expression, m.varnames, floats)
				if err == nil {
					setup.ConditionResults[last].Results[i] = res
				} else {
					fmt.Print(err.Error())
				}
			}
		}
	}

	m.trimLocked()
}

func (m *MultiSeriesRing) trimLocked() {
	points := len(m.times)
	if points <= m.maxPoints {
		return
	}

	start := points - m.maxPoints
	m.times = m.times[start:]

	for i := 0; i < m.vars; i++ {
		m.values[i] = m.values[i][start:]
	}

	theFilter := m.owner.GetFilter()
	if theFilter != nil {
		setup := theFilter.Setup
		if setup != nil && len(setup.ConditionResults) > start {
			setup.ConditionResults = setup.ConditionResults[start:]
		}
	}
}

func (m *MultiSeriesRing) ReadWindow(start, end int) ([]time.Time, [][]int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	n := len(m.times)
	if n == 0 {
		return nil, nil
	}

	if start < 0 {
		start = 0
	}
	if end <= 0 || end > n {
		end = n
	}
	if start >= end {
		return nil, nil
	}

	theFilter := m.owner.GetFilter()
	if theFilter != nil && theFilter.Enabled {
		return m.readWindowWithFilter(start, end)
	}

	// copy times
	times := make([]time.Time, end-start)
	copy(times, m.times[start:end])

	// copy values
	values := make([][]int, m.vars)
	for i := 0; i < m.vars; i++ {
		values[i] = make([]int, end-start)
		copy(values[i], m.values[i][start:end])
	}

	return times, values
}

func (m *MultiSeriesRing) readWindowWithFilter(start, end int) ([]time.Time, [][]int) {
	var times []time.Time
	values := make([][]int, m.vars)

	theFilter := m.owner.GetFilter()
	mode := theFilter.Setup.Mode
	components := theFilter.Setup.Components
	conditionResults := theFilter.Setup.ConditionResults

	for i := start; i < end; i++ {
		overallShown := mode == models.FilterModeAND // AND starts true, OR starts false

		for j := range components {
			c := components[j]
			if !c.Enabled {
				continue
			}
			res := conditionResults[i].Results[j]
			if mode == models.FilterModeAND {
				if !res {
					overallShown = false
					break
				}
			} else if mode == models.FilterModeOR {
				if res {
					overallShown = true
					break
				}
			}
		}

		if overallShown {
			for v := range m.vars {
				values[v] = append(values[v], m.values[v][i])
			}
			times = append(times, m.times[i])
		}
	}

	return times, values
}
