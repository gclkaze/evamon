package models

type FilterComponentChangeType string

const (
	FilterComponentAdded   FilterComponentChangeType = "ADDED"
	FilterComponentRemoved FilterComponentChangeType = "REMOVED"
	FilterComponentUpdated FilterComponentChangeType = "UPDATED"
)

type FilterComponentChange struct {
	Type      FilterComponentChangeType
	Component FilterComponent

	ExpressionChanged bool
	LabelChanged      bool
	EnabledChanged    bool
}

func (ch *FilterComponentChange) CopyFilterComponent() *FilterComponent {
	return ch.Component.Copy()
}

func DiffFilterComponents(old, new []FilterComponent) []FilterComponentChange {
	var changes []FilterComponentChange

	oldMap := make(map[string]FilterComponent)
	for _, c := range old {
		oldMap[c.ID] = c
	}

	newMap := make(map[string]FilterComponent)
	for _, c := range new {
		newMap[c.ID] = c
	}

	// added or updated
	for _, c := range new {
		oldC, exists := oldMap[c.ID]
		if !exists {
			changes = append(changes, FilterComponentChange{
				Type:      FilterComponentAdded,
				Component: c,
			})
		} else {
			exprChanged := oldC.Expression != c.Expression
			labelChanged := oldC.Label != c.Label
			enabledChanged := oldC.Enabled != c.Enabled

			if exprChanged || labelChanged || enabledChanged {
				changes = append(changes, FilterComponentChange{
					Type:              FilterComponentUpdated,
					Component:         c,
					ExpressionChanged: exprChanged,
					LabelChanged:      labelChanged,
					EnabledChanged:    enabledChanged,
				})
			}
		}
	}

	// removed
	for _, c := range old {
		if _, exists := newMap[c.ID]; !exists {
			changes = append(changes, FilterComponentChange{
				Type:      FilterComponentRemoved,
				Component: c,
			})
		}
	}

	return changes
}
