package models

import (
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

func GetFilterLabel(label, expression string) string {
	if strings.TrimSpace(label) == "" {
		return expression
	}
	return label
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
