package fynerenderer

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

func buildTab1(state *OperationsModalState) fyne.CanvasObject {
	leftList := widget.NewList(
		func() int { return len(state.AvailableComponents) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(state.AvailableComponents[id].Label)
		},
	)
	var rightList *widget.List

	rightList = widget.NewList(
		func() int { return len(state.SelectedItems) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				layout.NewSpacer(),
				widget.NewCheck("🔗 Maintain link", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			item := state.SelectedItems[id]
			row := obj.(*fyne.Container)
			label := row.Objects[0].(*widget.Label)
			check := row.Objects[2].(*widget.Check)

			label.SetText(item.ResolvedLabel(state.AvailableComponents))
			check.SetChecked(item.MaintainLink)
			check.OnChanged = func(checked bool) {
				item.MaintainLink = checked
				if !checked {
					item.BreakLink(state.AvailableComponents)
				}
				rightList.Refresh()
			}
		},
	)

	var selectedAvailable widget.ListItemID = -1
	var selectedChosen widget.ListItemID = -1

	leftList.OnSelected = func(id widget.ListItemID) {
		selectedAvailable = id
	}
	leftList.OnUnselected = func(id widget.ListItemID) {
		selectedAvailable = -1
	}

	rightList.OnSelected = func(id widget.ListItemID) {
		selectedChosen = id
	}
	rightList.OnUnselected = func(id widget.ListItemID) {
		selectedChosen = -1
	}

	addBtn := widget.NewButton("▶", func() {
		if selectedAvailable < 0 || selectedAvailable >= len(state.AvailableComponents) {
			return
		}
		c := state.AvailableComponents[selectedAvailable]

		// prevent duplicates
		for _, item := range state.SelectedItems {
			if item.SourceComponentID == c.ID {
				return
			}
		}

		state.SelectedItems = append(state.SelectedItems, models.NewFilterItem(c))
		rightList.Refresh()
	})

	removeBtn := widget.NewButton("◀", func() {
		if selectedChosen < 0 || selectedChosen >= len(state.SelectedItems) {
			return
		}

		item := state.SelectedItems[selectedChosen]

		if len(item.Files) > 0 {
			dialog.ShowConfirm(
				"Remove Constraint",
				fmt.Sprintf(
					"This constraint is assigned to %d file(s). Removing it will drop those assignments. Continue?",
					len(item.Files),
				),
				func(confirmed bool) {
					if confirmed {
						state.SelectedItems = append(
							state.SelectedItems[:selectedChosen],
							state.SelectedItems[selectedChosen+1:]...,
						)
						selectedChosen = -1
						rightList.Refresh()
					}
				},
				state.ParentWindow,
			)
			return
		}

		state.SelectedItems = append(
			state.SelectedItems[:selectedChosen],
			state.SelectedItems[selectedChosen+1:]...,
		)
		selectedChosen = -1
		rightList.Refresh()
	})

	buttons := container.NewVBox(
		layout.NewSpacer(),
		addBtn,
		removeBtn,
		layout.NewSpacer(),
	)

	return container.New(
		threeColLayout(200, 60, 200),
		container.NewBorder(widget.NewLabel("Available"), nil, nil, nil, leftList),
		buttons,
		container.NewBorder(widget.NewLabel("Selected"), nil, nil, nil, rightList),
	)
}
