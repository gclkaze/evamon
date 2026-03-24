package fynerenderer

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
	files []*models.ActionFile,
	onSave func([]models.FilterItem),
) {
	state := NewOperationsModalState(parent, components, files)

	tab1Content := buildTab1(state)
	tab2Content := buildTab2(state)

	tab1 := container.NewTabItem("Constraints", tab1Content)
	tab2 := container.NewTabItem("Assign to Files", tab2Content)

	tabs := container.NewAppTabs(tab1, tab2)
	tabs.OnSelected = func(ti *container.TabItem) {
		state.pruneAssignments()
		tab1.Content = buildTab1(state)
		tab2.Content = buildTab2(state)
		tabs.Refresh()
	}

	saveBtn := widget.NewButton("Save", func() {
		// break any unlinked items before returning
		for _, item := range state.SelectedItems {
			item.BreakLink(state.AvailableComponents)
		}

		result := make([]models.FilterItem, len(state.SelectedItems))
		for i, item := range state.SelectedItems {
			result[i] = *item
		}

		onSave(result)
	})

	cancelBtn := widget.NewButton("Cancel", func() {})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, buttons, nil, nil, tabs)

	d := dialog.NewCustom("Operations Setup", "Close", content, parent)
	d.Resize(fyne.NewSize(600, 400))
	d.Show()
}

/*func buildTab1(state *OperationsModalState) fyne.CanvasObject {
	return widget.NewLabel("Tab 1 - Constraints (coming next)")
}
*/
func buildTab2(state *OperationsModalState) fyne.CanvasObject {
	return widget.NewLabel("Tab 2 - Assign to Files (coming next)")
}
