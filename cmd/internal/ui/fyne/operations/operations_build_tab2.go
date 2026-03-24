package operations

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

func buildTab2(state *OperationsModalState) fyne.CanvasObject {

	var actionList *widget.List
	actionList = widget.NewList(
		func() int {
			if state.ActiveRule == nil {
				return 0
			}
			return len(state.ActiveRule.Actions)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				container.NewHBox(
					widget.NewButton("▲", nil),
					widget.NewButton("▼", nil),
				),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if state.ActiveRule == nil {
				return
			}
			action := state.ActiveRule.Actions[id]
			row := obj.(*fyne.Container)

			// with NewBorder: index 0 = center (label), index 1 = right (btnBox)
			label := row.Objects[0].(*widget.Label)
			btnBox := row.Objects[1].(*fyne.Container)
			upBtn := btnBox.Objects[0].(*widget.Button)
			downBtn := btnBox.Objects[1].(*widget.Button)

			label.SetText(fmt.Sprintf("%d. %s", id+1, action.Path))

			upBtn.OnTapped = func() {
				state.ActiveRule.MoveActionUp(id)
				actionList.Refresh()
			}
			downBtn.OnTapped = func() {
				state.ActiveRule.MoveActionDown(id)
				actionList.Refresh()
			}
		},
	)

	var selectedAction widget.ListItemID = -1

	actionList.OnSelected = func(id widget.ListItemID) {
		selectedAction = id
	}
	actionList.OnUnselected = func(id widget.ListItemID) {
		selectedAction = -1
	}

	// file picker helper
	openFilePicker := func(onPicked func(path string)) {
		fd := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			onPicked(uc.URI().Path())
		}, state.ParentWindow)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".eva"}))
		fd.Show()
	}

	addActionBtn := widget.NewButton("Add", func() {
		if state.ActiveRule == nil {
			return
		}
		openFilePicker(func(path string) {
			state.ActiveRule.AddAction(models.NewActionFile(path))
			actionList.Refresh()
		})
	})

	editActionBtn := widget.NewButton("Edit", func() {
		if state.ActiveRule == nil || selectedAction < 0 || selectedAction >= len(state.ActiveRule.Actions) {
			return
		}
		openFilePicker(func(path string) {
			state.ActiveRule.Actions[selectedAction].Path = path
			actionList.Refresh()
		})
	})

	removeActionBtn := widget.NewButton("Remove", func() {
		if state.ActiveRule == nil {
			return
		}
		if selectedAction < 0 || selectedAction >= len(state.ActiveRule.Actions) {
			return
		}
		state.ActiveRule.RemoveActionAt(selectedAction)
		if selectedAction >= len(state.ActiveRule.Actions) {
			selectedAction = len(state.ActiveRule.Actions) - 1
		}
		actionList.Refresh()
	})

	actionToolbar := container.NewHBox(addActionBtn, editActionBtn, removeActionBtn)

	// left — conditions list
	ruleList := widget.NewList(
		func() int { return len(state.SelectedRules) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			rule := state.SelectedRules[id]
			label := obj.(*widget.Label)

			text := rule.ResolvedLabel(state.AvailableComponents)
			if len(rule.Actions) > 0 {
				text = fmt.Sprintf("%s  (%d actions)", text, len(rule.Actions))
			}
			label.SetText(text)
		},
	)

	ruleList.OnSelected = func(id widget.ListItemID) {
		state.ActiveRule = state.SelectedRules[id]
		selectedAction = -1
		actionList.Refresh()
		ruleList.Refresh()
	}
	ruleList.OnUnselected = func(id widget.ListItemID) {
		state.ActiveRule = nil
		selectedAction = -1
		actionList.Refresh()
	}

	// restore selection if ActiveRule is already set
	if state.ActiveRule != nil {
		for i, rule := range state.SelectedRules {
			if rule == state.ActiveRule {
				ruleList.Select(widget.ListItemID(i))
				break
			}
		}
	}

	// center → button
	centerButtons := container.NewVBox(
		layout.NewSpacer(),
		widget.NewButton("→", func() {
			if state.ActiveRule == nil {
				return
			}
			openFilePicker(func(path string) {
				state.ActiveRule.AddAction(models.NewActionFile(path))
				actionList.Refresh()
				ruleList.Refresh()
			})
		}),
		layout.NewSpacer(),
	)

	actionPanel := container.NewBorder(
		container.NewVBox(widget.NewLabel("Actions"), actionToolbar),
		nil, nil, nil,
		actionList,
	)

	return container.New(
		threeColLayout(200, 60, 400),
		container.NewBorder(widget.NewLabel("Conditions"), nil, nil, nil, ruleList),
		centerButtons,
		actionPanel,
	)
}
