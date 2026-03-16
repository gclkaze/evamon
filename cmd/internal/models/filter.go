package models

import (
	"fmt"
	"strings"

	"github.com/gclkaze/evamon/pkg/utils"
)

type FilterMode string

const (
	FilterModeAND FilterMode = "AND"
	FilterModeOR  FilterMode = "OR"
)

type FilterComponent struct {
	ID         string `json:"id,omitempty"`
	Label      string `json:"label,omitempty"`
	Expression string `json:"expression"`
	Enabled    bool   `json:"enabled"`
}

type ConditionResult struct {
	Results map[string]bool `json:"-"`
}

type FilterSetup struct {
	Mode       FilterMode        `json:"mode"`
	Components []FilterComponent `json:"components"`
	//one for each time point
	ConditionResults []ConditionResult `json:"-"`
}

type Filter struct {
	Enabled bool         `json:"enabled"`
	Setup   *FilterSetup `json:"setup,omitempty"`
}

func NewFilterComponent(expression string) FilterComponent {
	return FilterComponent{
		ID:         utils.GetRandomString(),
		Expression: expression,
		Enabled:    true,
		Label:      "",
	}
}

func NewFilter() *Filter {
	f := &Filter{Enabled: true, Setup: NewFilterSetup()}
	return f
}

func NewFilterSetup() *FilterSetup {
	f := &FilterSetup{Mode: FilterModeAND, Components: make([]FilterComponent, 0), ConditionResults: make([]ConditionResult, 0)}
	return f
}

func (f *FilterComponent) Copy() *FilterComponent {
	return &FilterComponent{
		ID:         utils.GetRandomString(),
		Expression: f.Expression,
		Enabled:    f.Enabled,
		Label:      f.Label,
	}
}

/*func GetFilterLabel(label, expression string) string {
	if strings.TrimSpace(label) == "" {
		return expression
	}
	return label
}*/

const maxFilterLabelLength = 20

func GetFilterLabel(label, expression string, index int) string {
	l := strings.TrimSpace(label)
	if l == "" {
		return fmt.Sprintf("Filter#%d", index+1)
	}
	if len(l) > maxFilterLabelLength {
		return l[:maxFilterLabelLength] + "…"
	}
	return l
}

func (f *Filter) HasEnabledComponents() bool {
	if f == nil || f.Setup == nil {
		return false
	}

	for _, c := range f.Setup.Components {
		if c.Enabled {
			return true
		}
	}

	return false
}
func (s *FilterSetup) InitializeFilter(maxPoints int) {
	s.ConditionResults = make([]ConditionResult, 0, maxPoints)
}

func (c *FilterComponent) CopyFrom(from *FilterComponent) {
	c.Enabled = from.Enabled
	c.Expression = from.Expression
	c.Label = from.Label
}
