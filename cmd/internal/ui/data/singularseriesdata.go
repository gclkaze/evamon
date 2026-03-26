package data

import (
	"sync"
	"time"

	"github.com/gclkaze/evamon/cmd/internal/models"
	port "github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/pkg/utils"
)

// SingularSeriesData implements IMultiSeriesData for a single boolean variable.
// It retains only the most-recent value and its timestamp.
// Trigger-rule evaluation runs on every Append call, identical to MultiSeriesRing.
type SingularSeriesData struct {
	mu sync.RWMutex

	hasValue  bool
	lastValue int
	lastTime  time.Time

	owner   port.IDiagram
	varname string // "$varname" form, used by RunExpression

	triggerSender TriggerSendFunc
	lastFired     map[string]bool // OriginalComponentID → last evaluation result
}

func NewSingularSeriesData(owner port.IDiagram, varname string) *SingularSeriesData {
	if varname != "" && varname[0] != '$' {
		varname = "$" + varname
	}
	return &SingularSeriesData{
		owner:   owner,
		varname: varname,
	}
}

func (s *SingularSeriesData) Vars() int     { return 1 }
func (s *SingularSeriesData) MaxPoints() int { return 1 }

func (s *SingularSeriesData) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.hasValue {
		return 1
	}
	return 0
}

func (s *SingularSeriesData) Append(at time.Time, nums []int) {
	if len(nums) != 1 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.hasValue = true
	s.lastValue = nums[0]
	s.lastTime = at

	if s.triggerSender != nil {
		s.evaluateTriggerRules([]float64{float64(nums[0])})
	}
}

func (s *SingularSeriesData) ReadWindow(start, end int) ([]time.Time, [][]int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.hasValue || start > 0 || end <= 0 {
		return nil, nil
	}

	return []time.Time{s.lastTime}, [][]int{{s.lastValue}}
}

func (s *SingularSeriesData) ApplyFilterChanges(_ []models.FilterComponentChange) {}

func (s *SingularSeriesData) SetTriggerSender(fn TriggerSendFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.triggerSender = fn
}

// evaluateTriggerRules runs inside the existing mu.Lock in Append.
func (s *SingularSeriesData) evaluateTriggerRules(floats []float64) {
	if s.owner == nil {
		return
	}

	rules := s.owner.GetTriggerRules()
	if len(rules) == 0 {
		return
	}

	var filterComponents []models.FilterComponent
	if f := s.owner.GetFilter(); f != nil && f.Setup != nil {
		filterComponents = f.Setup.Components
	}

	if s.lastFired == nil {
		s.lastFired = make(map[string]bool)
	}

	sender := s.triggerSender

	for _, rule := range rules {
		if len(rule.Actions) == 0 {
			continue
		}

		expr := rule.ResolvedExpression(filterComponents)
		if expr == "" {
			continue
		}

		res, err := utils.RunExpression(expr, []string{s.varname}, floats)
		if err != nil {
			continue
		}

		id := rule.OriginalComponentID
		prev := s.lastFired[id]
		s.lastFired[id] = res

		if !res || prev {
			continue
		}

		files := make([]string, len(rule.Actions))
		for i, a := range rule.Actions {
			files[i] = a.Path
		}
		ruleID := id
		go sender(ruleID, files)
	}
}

var _ IMultiSeriesData = (*SingularSeriesData)(nil)
