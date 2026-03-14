package data

import "github.com/gclkaze/evamon/cmd/internal/models"

type FilterComponentChangeType string

const (
	FilterComponentAdded   FilterComponentChangeType = "ADDED"
	FilterComponentRemoved FilterComponentChangeType = "REMOVED"
	FilterComponentUpdated FilterComponentChangeType = "UPDATED"
)

type FilterComponentChange struct {
	Type      FilterComponentChangeType
	Component models.FilterComponent
}

func DiffFilterComponents(old, new []models.FilterComponent) []FilterComponentChange {
	var changes []FilterComponentChange

	oldMap := make(map[string]models.FilterComponent)
	for _, c := range old {
		oldMap[c.ID] = c
	}

	newMap := make(map[string]models.FilterComponent)
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
		} else if oldC.Expression != c.Expression {
			// only expression change triggers recalculation
			// enabled/disabled change is intentionally ignored
			changes = append(changes, FilterComponentChange{
				Type:      FilterComponentUpdated,
				Component: c,
			})
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
