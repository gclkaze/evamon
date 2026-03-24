package fynerenderer

import (
	"fyne.io/fyne/v2"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

type OperationsModalState struct {
	ParentWindow        fyne.Window
	AvailableComponents []models.FilterComponent
	SelectedRules       []*models.TriggerRule
	ActiveRule          *models.TriggerRule
}

func NewOperationsModalState(
	parent fyne.Window,
	components []models.FilterComponent,
) *OperationsModalState {
	return &OperationsModalState{
		ParentWindow:        parent,
		AvailableComponents: components,
		SelectedRules:       make([]*models.TriggerRule, 0),
		ActiveRule:          nil,
	}
}

func (ms *OperationsModalState) pruneAssignments() {
	// nothing to prune at state level — actions are owned by each TriggerRule
}

func (ms *OperationsModalState) isFileAssigned(file *models.ActionFile, item *models.FilterItem) bool {
	return item.HasFile(file.Path)
}
