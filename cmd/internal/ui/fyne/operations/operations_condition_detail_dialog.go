package operations

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

func showConditionDetailDialog(
	rule *models.TriggerRule,
	state *OperationsModalState,
	onUpdated func(),
	parent fyne.Window,
) {
	labelEntry := buildLabelEntry(rule)
	expressionEntry := buildExpressionEntry(rule)

	var d dialog.Dialog

	saveBtn := buildConditionSaveBtn(rule, state, labelEntry, expressionEntry, onUpdated, &d)
	cancelBtn := widget.NewButton("Cancel", func() { d.Hide() })

	content := buildConditionDialogContent(labelEntry, expressionEntry, saveBtn, cancelBtn)

	d = dialog.NewCustomWithoutButtons("Edit Condition", content, parent)
	d.Resize(fyne.NewSize(420, 260))
	d.Show()
}

func buildLabelEntry(rule *models.TriggerRule) *widget.Entry {
	e := widget.NewEntry()
	e.SetText(rule.Label)
	return e
}

func buildExpressionEntry(rule *models.TriggerRule) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.SetText(rule.Expression)
	return e
}

func buildConditionSaveBtn(
	rule *models.TriggerRule,
	state *OperationsModalState,
	labelEntry *widget.Entry,
	expressionEntry *widget.Entry,
	onUpdated func(),
	d *dialog.Dialog,
) *widget.Button {
	btn := widget.NewButton("Save", func() {
		rule.Label = labelEntry.Text
		rule.Expression = expressionEntry.Text
		rule.Edited = true
		state.Differentiator.OnLabelChanged(state.SelectedRules)
		state.Differentiator.OnExpressionChanged(state.SelectedRules)
		state.NotifyChanged()
		onUpdated()
		(*d).Hide()
	})
	btn.Importance = widget.HighImportance
	return btn
}

func buildConditionDialogContent(
	labelEntry *widget.Entry,
	expressionEntry *widget.Entry,
	saveBtn *widget.Button,
	cancelBtn *widget.Button,
) fyne.CanvasObject {
	form := widget.NewForm(
		widget.NewFormItem("Label", labelEntry),
		widget.NewFormItem("Expression", expressionEntry),
	)
	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn, layout.NewSpacer())
	return container.NewBorder(nil, buttons, nil, nil, form)
}
