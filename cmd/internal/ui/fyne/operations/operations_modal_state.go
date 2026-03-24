package operations

import (
	"fyne.io/fyne/v2"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

type OperationsModalState struct {
	ParentWindow        fyne.Window
	AvailableComponents []models.FilterComponent
	SelectedRules       []*models.TriggerRule
	ActiveRule          *models.TriggerRule
	Differentiator      *models.OperationsStateDifferentiator
	OnChanged           func()
}

func NewOperationsModalState(
	parent fyne.Window,
	components []models.FilterComponent,
	initialRules []*models.TriggerRule,
) *OperationsModalState {
	return &OperationsModalState{
		ParentWindow:        parent,
		AvailableComponents: components,
		SelectedRules:       initialRules,
		ActiveRule:          nil,
		Differentiator:      models.NewOperationsStateDifferentiator(initialRules),
		OnChanged:           func() {},
	}
}

func (ms *OperationsModalState) pruneAssignments() {}

func (ms *OperationsModalState) NotifyChanged() {
	if ms.OnChanged != nil {
		ms.OnChanged()
	}
}
