package operations

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

type OperationsConditionTab struct {
	state         *OperationsModalState
	leftList      *widget.List
	rightList     *widget.List
	selectedLeft  int
	selectedRight int
	validator     ExpressionValidator
}

func NewOperationsConditionTab(state *OperationsModalState, validator ExpressionValidator) *OperationsConditionTab {
	return &OperationsConditionTab{
		state:         state,
		selectedLeft:  -1,
		selectedRight: -1,
		validator:     validator,
	}
}

func (t *OperationsConditionTab) Build() fyne.CanvasObject {
	t.leftList = t.buildLeftList()
	t.rightList = t.buildRightList()
	buttons := t.buildButtons()

	rightPanel := container.NewBorder(
		container.NewVBox(widget.NewLabel("Active Conditions"), t.newRuleBtn()),
		nil, nil, nil,
		t.rightList,
	)

	return container.New(
		threeColLayout(200, 60, 350),
		container.NewBorder(widget.NewLabel("Available Conditions"), nil, nil, nil, t.leftList),
		buttons,
		rightPanel,
	)
}

func (t *OperationsConditionTab) buildLeftList() *widget.List {
	l := widget.NewList(
		func() int { return len(t.state.AvailableComponents) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(t.state.AvailableComponents[id].Label)
		},
	)
	l.OnSelected = func(id widget.ListItemID) { t.selectedLeft = id }
	l.OnUnselected = func(id widget.ListItemID) { t.selectedLeft = -1 }
	return l
}

func (t *OperationsConditionTab) buildRightList() *widget.List {
	var l *widget.List
	l = widget.NewList(
		func() int { return len(t.state.SelectedRules) },
		func() fyne.CanvasObject { return t.buildRightListRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { t.updateRightListRow(id, obj, l) },
	)
	l.OnSelected = func(id widget.ListItemID) { t.selectedRight = id }
	l.OnUnselected = func(id widget.ListItemID) { t.selectedRight = -1 }
	return l
}

func (t *OperationsConditionTab) buildRightListRow() fyne.CanvasObject {
	return container.NewBorder(
		nil, nil, nil,
		container.NewHBox(
			widget.NewCheck("🔗 Maintain link", nil),
			widget.NewButton("Edit", nil),
		),
		widget.NewLabel(""),
	)
}

func (t *OperationsConditionTab) updateRightListRow(id widget.ListItemID, obj fyne.CanvasObject, l *widget.List) {
	rule := t.state.SelectedRules[id]
	row := obj.(*fyne.Container)
	label := row.Objects[0].(*widget.Label)
	btnBox := row.Objects[1].(*fyne.Container)
	check := btnBox.Objects[0].(*widget.Check)
	editBtn := btnBox.Objects[1].(*widget.Button)

	label.SetText(rule.ResolvedLabel(t.state.AvailableComponents))
	t.updateCheckState(rule, check, editBtn, l)
	t.updateEditBtn(rule, editBtn)
}

func (t *OperationsConditionTab) updateCheckState(rule *models.TriggerRule, check *widget.Check, editBtn *widget.Button, l *widget.List) {
	check.SetChecked(rule.MaintainLink)
	if rule.Edited {
		check.Disable()
	} else {
		check.Enable()
	}
	check.OnChanged = func(checked bool) {
		t.toggleMaintainLink(rule, checked, editBtn)
		l.Refresh()
	}
}

func (t *OperationsConditionTab) updateEditBtn(rule *models.TriggerRule, editBtn *widget.Button) {
	if rule.MaintainLink {
		editBtn.Disable()
	} else {
		editBtn.Enable()
	}
	editBtn.OnTapped = func() {
		t.openConditionDetail(rule)
	}
}

func (t *OperationsConditionTab) buildButtons() *fyne.Container {
	addBtn := widget.NewButton("▶", func() { t.addRule() })
	removeBtn := widget.NewButton("◀", func() { t.removeRule() })
	return container.NewVBox(layout.NewSpacer(), addBtn, removeBtn, layout.NewSpacer())
}

func (t *OperationsConditionTab) addRule() {
	if t.selectedLeft < 0 || t.selectedLeft >= len(t.state.AvailableComponents) {
		return
	}
	c := t.state.AvailableComponents[t.selectedLeft]
	for _, rule := range t.state.SelectedRules {
		if rule.SourceComponentID == c.ID {
			return
		}
	}
	t.state.SelectedRules = append(t.state.SelectedRules, models.NewTriggerRule(c))
	t.state.Differentiator.OnRuleAdded(t.state.SelectedRules)
	t.state.NotifyChanged()
	t.rightList.Refresh()
}

func (t *OperationsConditionTab) removeRule() {
	if t.selectedRight < 0 || t.selectedRight >= len(t.state.SelectedRules) {
		return
	}
	rule := t.state.SelectedRules[t.selectedRight]
	if len(rule.Actions) > 0 {
		t.confirmRemoveRule(rule)
		return
	}
	t.doRemoveRule()
}

func (t *OperationsConditionTab) confirmRemoveRule(rule *models.TriggerRule) {
	dialog.ShowConfirm(
		"Remove Condition",
		fmt.Sprintf("This condition has %d action(s) assigned. Removing it will drop those actions. Continue?", len(rule.Actions)),
		func(confirmed bool) {
			if confirmed {
				t.doRemoveRule()
			}
		},
		t.state.ParentWindow,
	)
}

func (t *OperationsConditionTab) doRemoveRule() {
	rule := t.state.SelectedRules[t.selectedRight]
	t.state.SelectedRules = append(t.state.SelectedRules[:t.selectedRight], t.state.SelectedRules[t.selectedRight+1:]...)
	if t.state.ActiveRule == rule {
		t.state.ActiveRule = nil
	}
	t.state.Differentiator.OnRuleRemoved(t.state.SelectedRules)
	t.state.NotifyChanged()
	t.selectedRight = -1
	t.rightList.Refresh()
}

func (t *OperationsConditionTab) toggleMaintainLink(rule *models.TriggerRule, checked bool, editBtn *widget.Button) {
	rule.MaintainLink = checked
	if !checked {
		rule.BreakLink(t.state.AvailableComponents)
		editBtn.Enable()
	} else {
		rule.RestoreLink()
		editBtn.Disable()
	}
	t.state.Differentiator.OnMaintainLinkChanged(t.state.SelectedRules)
	t.state.NotifyChanged()
}

func (t *OperationsConditionTab) openConditionDetail(rule *models.TriggerRule) {
	showConditionDetailDialog(rule, t.state, t.validator, func(updated *models.TriggerRule) {
		t.rightList.Refresh()
	}, t.state.ParentWindow)
}

func (t *OperationsConditionTab) newRuleBtn() *widget.Button {
	return widget.NewButton("+ New", func() {
		t.openNewConditionDetail()
	})
}

func (t *OperationsConditionTab) openNewConditionDetail() {
	showConditionDetailDialog(nil, t.state, t.validator, func(rule *models.TriggerRule) {
		if strings.TrimSpace(rule.Label) == "" && strings.TrimSpace(rule.Expression) == "" {
			return
		}
		t.state.SelectedRules = append(t.state.SelectedRules, rule)
		t.state.Differentiator.OnRuleAdded(t.state.SelectedRules)
		t.state.NotifyChanged()
		t.rightList.Refresh()
	}, t.state.ParentWindow)
}
