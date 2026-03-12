package data

import (
	"sync"
	"time"

	port "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type MultiSeriesRing struct {
	mu sync.RWMutex

	maxPoints int
	vars      int

	times  []time.Time
	values [][]int

	owner port.IDiagram
}

func NewMultiSeriesRing(maxPoints int, vars int, owner port.IDiagram) *MultiSeriesRing {
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
	}

	for i := 0; i < vars; i++ {
		ms.values[i] = make([]int, 0, maxPoints)
	}

	return ms
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
