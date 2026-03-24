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
		func() int { return len(state.SelectedRules) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				container.NewHBox(
					widget.NewCheck("🔗 Maintain link", nil),
					widget.NewButton("Edit", nil),
				),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			rule := state.SelectedRules[id]
			row := obj.(*fyne.Container)

			label := row.Objects[0].(*widget.Label)
			btnBox := row.Objects[1].(*fyne.Container)
			check := btnBox.Objects[0].(*widget.Check)
			editBtn := btnBox.Objects[1].(*widget.Button)

			label.SetText(rule.ResolvedLabel(state.AvailableComponents))

			check.SetChecked(rule.MaintainLink)

			if rule.Edited {
				check.Disable()
			} else {
				check.Enable()
			}

			check.OnChanged = func(checked bool) {
				rule.MaintainLink = checked
				if !checked {
					rule.BreakLink(state.AvailableComponents)
					editBtn.Enable()
				} else {
					rule.RestoreLink()
					editBtn.Disable()
				}
				rightList.Refresh()
			}

			if rule.MaintainLink {
				editBtn.Disable()
			} else {
				editBtn.Enable()
			}

			editBtn.OnTapped = func() {
				showConditionDetailDialog(rule, state.AvailableComponents, func() {
					rightList.Refresh()
				}, state.ParentWindow)
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

		for _, rule := range state.SelectedRules {
			if rule.SourceComponentID == c.ID {
				return
			}
		}

		state.SelectedRules = append(state.SelectedRules, models.NewTriggerRule(c))
		rightList.Refresh()
	})

	removeBtn := widget.NewButton("◀", func() {
		if selectedChosen < 0 || selectedChosen >= len(state.SelectedRules) {
			return
		}

		rule := state.SelectedRules[selectedChosen]

		if len(rule.Actions) > 0 {
			dialog.ShowConfirm(
				"Remove Condition",
				fmt.Sprintf(
					"This condition has %d action(s) assigned. Removing it will drop those actions. Continue?",
					len(rule.Actions),
				),
				func(confirmed bool) {
					if confirmed {
						state.SelectedRules = append(
							state.SelectedRules[:selectedChosen],
							state.SelectedRules[selectedChosen+1:]...,
						)
						if state.ActiveRule != nil && state.ActiveRule == rule {
							state.ActiveRule = nil
						}
						selectedChosen = -1
						rightList.Refresh()
					}
				},
				state.ParentWindow,
			)
			return
		}

		state.SelectedRules = append(
			state.SelectedRules[:selectedChosen],
			state.SelectedRules[selectedChosen+1:]...,
		)
		if state.ActiveRule != nil && state.ActiveRule == rule {
			state.ActiveRule = nil
		}
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
		threeColLayout(200, 60, 350),
		container.NewBorder(widget.NewLabel("Available Conditions"), nil, nil, nil, leftList),
		buttons,
		container.NewBorder(widget.NewLabel("Active Conditions"), nil, nil, nil, rightList),
	)
}

func showConditionDetailDialog(rule *models.TriggerRule, components []models.FilterComponent, onUpdated func(), parent fyne.Window) {
	labelEntry := widget.NewEntry()
	labelEntry.SetText(rule.Label)

	expressionEntry := widget.NewMultiLineEntry()
	expressionEntry.SetText(rule.Expression)

	form := widget.NewForm(
		widget.NewFormItem("Label", labelEntry),
		widget.NewFormItem("Expression", expressionEntry),
	)

	var d dialog.Dialog

	saveBtn := widget.NewButton("Save", func() {
		rule.Label = labelEntry.Text
		rule.Expression = expressionEntry.Text
		rule.Edited = true
		onUpdated()
		d.Hide()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn, layout.NewSpacer())
	content := container.NewBorder(nil, buttons, nil, nil, form)

	d = dialog.NewCustomWithoutButtons("Edit Condition", content, parent)
	d.Resize(fyne.NewSize(420, 260))
	d.Show()
}
