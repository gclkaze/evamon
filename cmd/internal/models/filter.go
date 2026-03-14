package models

type FilterMode string

const (
	FilterModeAND FilterMode = "AND"
	FilterModeOR  FilterMode = "OR"
)

type FilterComponent struct {
	ID         string `json:"id,omitempty"`
	Expression string `json:"expression"`
	Enabled    bool   `json:"enabled"`
}

type ConditionResult struct {
	Results []bool `json:"-"`
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
