package operations

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

func ShowOperationsModal(
	parent fyne.Window,
	components []models.FilterComponent,
	onSave func([]models.TriggerRule),
) {
	state := NewOperationsModalState(parent, components)

	tab1 := container.NewTabItem("Conditions", buildTab1(state))
	tab2 := container.NewTabItem("Actions", buildTab2(state))

	tabs := container.NewAppTabs(tab1, tab2)

	var d dialog.Dialog

	tabs.OnSelected = func(ti *container.TabItem) {
		state.pruneAssignments()
		tab1.Content = buildTab1(state)
		tab2.Content = buildTab2(state)
		tabs.Refresh()
	}

	saveBtn := widget.NewButton("Save", func() {
		for _, rule := range state.SelectedRules {
			rule.BreakLink(state.AvailableComponents)
		}

		result := make([]models.TriggerRule, len(state.SelectedRules))
		for i, rule := range state.SelectedRules {
			result[i] = *rule
		}

		onSave(result)
		d.Hide()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, buttons, nil, nil, tabs)

	d = dialog.NewCustomWithoutButtons("Operations Setup", content, parent)
	d.Resize(fyne.NewSize(900, 660))
	d.Show()
}
