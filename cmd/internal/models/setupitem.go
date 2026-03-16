package models

// SetupItem is one variable/series to visualize.
type SetupItem struct {
	Variable     string       `json:"variable"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	VariableType ValueType    `json:"variableType"`
	WindowStyle  *WindowStyle `json:"windowStyle,omitempty"`

	// Typed union:
	// - for boolean diagrams: BooleanStyle
	// - for bar diagrams:     BarStyle
	DiagramStyle Style `json:"diagramStyle,omitempty"`

	Filter *Filter `json:"filter,omitempty"`

	MultiVariableSetup []MultiSetupItem `json:"multiVariableSetup,omitempty"`
}

func (s SetupItem) HasFilter() bool {
	return s.Filter != nil &&
		s.Filter.Setup != nil &&
		len(s.Filter.Setup.Components) > 0
}
func (s SetupItem) IsFilterEnabled() bool {
	if s.Filter == nil {
		return false
	}
	return s.Filter.Enabled
}
func (s SetupItem) GetFilterMode() FilterMode {
	if s.Filter == nil || s.Filter.Setup == nil {
		return FilterModeAND
	}
	return s.Filter.Setup.Mode
}

func (s SetupItem) GetFilter() *Filter {
	return s.Filter
}

func (s *SetupItem) HandleFilterChange(new []FilterComponent, changes []FilterComponentChange, newFilter *Filter) {
	for i := range changes {
		change := changes[i]
		switch change.Type {
		case FilterComponentRemoved:
			{
				id := change.Component.ID
				for i := range s.Filter.Setup.Components {
					if s.Filter.Setup.Components[i].ID == id {
						s.Filter.Setup.Components = append(s.Filter.Setup.Components[:i], s.Filter.Setup.Components[i+1:]...)
						break
					}
				}
			}
		case FilterComponentAdded:
			{
				newComponent := change.Component.Copy()
				if s.Filter == nil {
					s.Filter = NewFilter()
				}
				s.Filter.Setup.Components = append(s.Filter.Setup.Components, *newComponent)
			}
		case FilterComponentUpdated:
			{
				id := change.Component.ID
				for i := range s.Filter.Setup.Components {
					if s.Filter.Setup.Components[i].ID == id {
						s.Filter.Setup.Components[i].CopyFrom(&change.Component)
						break
					}
				}
			}
		default:
		}
	}
}
