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

type OperationsActionsTab struct {
	state          *OperationsModalState
	ruleList       *widget.List
	actionList     *widget.List
	selectedAction int
}

func NewOperationsActionsTab(state *OperationsModalState) *OperationsActionsTab {
	return &OperationsActionsTab{
		state:          state,
		selectedAction: -1,
	}
}

func (t *OperationsActionsTab) Build() fyne.CanvasObject {
	t.actionList = t.buildActionList()
	t.ruleList = t.buildRuleList()

	return container.New(
		threeColLayout(200, 60, 400),
		container.NewBorder(widget.NewLabel("Conditions"), nil, nil, nil, t.ruleList),
		t.buildCenterButtons(),
		container.NewBorder(
			container.NewVBox(widget.NewLabel("Actions"), t.buildActionToolbar()),
			nil, nil, nil,
			t.actionList,
		),
	)
}

func (t *OperationsActionsTab) buildRuleList() *widget.List {
	l := widget.NewList(
		func() int { return len(t.state.SelectedRules) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			t.updateRuleListRow(id, obj)
		},
	)
	l.OnSelected = func(id widget.ListItemID) {
		t.state.ActiveRule = t.state.SelectedRules[id]
		t.selectedAction = -1
		t.actionList.Refresh()
		l.Refresh()
	}
	l.OnUnselected = func(id widget.ListItemID) {
		t.state.ActiveRule = nil
		t.selectedAction = -1
		t.actionList.Refresh()
	}
	t.restoreRuleSelection(l)
	return l
}

func (t *OperationsActionsTab) updateRuleListRow(id widget.ListItemID, obj fyne.CanvasObject) {
	rule := t.state.SelectedRules[id]
	label := obj.(*widget.Label)
	text := rule.ResolvedLabel(t.state.AvailableComponents)
	if len(rule.Actions) > 0 {
		text = fmt.Sprintf("%s  (%d actions)", text, len(rule.Actions))
	}
	label.SetText(text)
}

func (t *OperationsActionsTab) restoreRuleSelection(l *widget.List) {
	if t.state.ActiveRule == nil {
		return
	}
	for i, rule := range t.state.SelectedRules {
		if rule == t.state.ActiveRule {
			l.Select(widget.ListItemID(i))
			break
		}
	}
}

func (t *OperationsActionsTab) buildActionList() *widget.List {
	var l *widget.List
	l = widget.NewList(
		func() int {
			if t.state.ActiveRule == nil {
				return 0
			}
			return len(t.state.ActiveRule.Actions)
		},
		func() fyne.CanvasObject { return t.buildActionListRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { t.updateActionListRow(id, obj, l) },
	)
	l.OnSelected = func(id widget.ListItemID) { t.selectedAction = id }
	l.OnUnselected = func(id widget.ListItemID) { t.selectedAction = -1 }
	return l
}

func (t *OperationsActionsTab) buildActionListRow() fyne.CanvasObject {
	return container.NewBorder(
		nil, nil, nil,
		container.NewHBox(
			widget.NewButton("▲", nil),
			widget.NewButton("▼", nil),
		),
		widget.NewLabel(""),
	)
}

func (t *OperationsActionsTab) updateActionListRow(id widget.ListItemID, obj fyne.CanvasObject, l *widget.List) {
	if t.state.ActiveRule == nil {
		return
	}
	action := t.state.ActiveRule.Actions[id]
	row := obj.(*fyne.Container)
	label := row.Objects[0].(*widget.Label)
	btnBox := row.Objects[1].(*fyne.Container)
	upBtn := btnBox.Objects[0].(*widget.Button)
	downBtn := btnBox.Objects[1].(*widget.Button)

	label.SetText(fmt.Sprintf("%d. %s", id+1, action.Path))
	upBtn.OnTapped = func() { t.moveActionUp(id, l) }
	downBtn.OnTapped = func() { t.moveActionDown(id, l) }
}

func (t *OperationsActionsTab) buildActionToolbar() *fyne.Container {
	return container.NewHBox(
		widget.NewButton("Add", func() { t.openFilePicker(func(path string) { t.addAction(path) }) }),
		widget.NewButton("Edit", func() { t.openFilePicker(func(path string) { t.editAction(path) }) }),
		widget.NewButton("Remove", func() { t.removeAction() }),
	)
}

func (t *OperationsActionsTab) buildCenterButtons() *fyne.Container {
	return container.NewVBox(
		layout.NewSpacer(),
		widget.NewButton("→", func() {
			t.openFilePicker(func(path string) { t.addAction(path) })
		}),
		layout.NewSpacer(),
	)
}

func (t *OperationsActionsTab) openFilePicker(onPicked func(string)) {
	fd := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		onPicked(uc.URI().Path())
	}, t.state.ParentWindow)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".eva"}))
	fd.Show()
}

func (t *OperationsActionsTab) addAction(path string) {
	if t.state.ActiveRule == nil {
		return
	}
	t.state.ActiveRule.AddAction(models.NewActionFile(path))
	t.state.Differentiator.OnActionsChanged(t.state.SelectedRules)
	t.state.NotifyChanged()
	t.actionList.Refresh()
	t.ruleList.Refresh()
}

func (t *OperationsActionsTab) editAction(path string) {
	if t.state.ActiveRule == nil || t.selectedAction < 0 || t.selectedAction >= len(t.state.ActiveRule.Actions) {
		return
	}
	t.state.ActiveRule.Actions[t.selectedAction].Path = path
	t.state.Differentiator.OnActionsChanged(t.state.SelectedRules)
	t.state.NotifyChanged()
	t.actionList.Refresh()
}

func (t *OperationsActionsTab) removeAction() {
	if t.state.ActiveRule == nil || t.selectedAction < 0 || t.selectedAction >= len(t.state.ActiveRule.Actions) {
		return
	}
	t.state.ActiveRule.RemoveActionAt(t.selectedAction)
	if t.selectedAction >= len(t.state.ActiveRule.Actions) {
		t.selectedAction = len(t.state.ActiveRule.Actions) - 1
	}
	t.state.Differentiator.OnActionsChanged(t.state.SelectedRules)
	t.state.NotifyChanged()
	t.actionList.Refresh()
}

func (t *OperationsActionsTab) moveActionUp(index int, l *widget.List) {
	if t.state.ActiveRule == nil {
		return
	}
	t.state.ActiveRule.MoveActionUp(index)
	t.state.Differentiator.OnActionsChanged(t.state.SelectedRules)
	t.state.NotifyChanged()
	l.Refresh()
}

func (t *OperationsActionsTab) moveActionDown(index int, l *widget.List) {
	if t.state.ActiveRule == nil {
		return
	}
	t.state.ActiveRule.MoveActionDown(index)
	t.state.Differentiator.OnActionsChanged(t.state.SelectedRules)
	t.state.NotifyChanged()
	l.Refresh()
}
