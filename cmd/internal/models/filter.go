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

type FilterSetup struct {
	Mode       FilterMode        `json:"mode"`
	Components []FilterComponent `json:"components"`
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
