package operations

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

type ExpressionValidator func(expression string) error

type conditionErrorBox struct {
	errorText *canvas.Text
	errorIcon *widget.Icon
	box       *fyne.Container
}

func newConditionErrorBox() *conditionErrorBox {
	errorText := canvas.NewText("", color.NRGBA{R: 211, G: 47, B: 47, A: 255})
	errorText.TextSize = 12
	errorIcon := widget.NewIcon(theme.WarningIcon())
	box := container.NewHBox(errorIcon, errorText)
	box.Hide()
	return &conditionErrorBox{errorText: errorText, errorIcon: errorIcon, box: box}
}

func (e *conditionErrorBox) show(msg string) {
	e.errorText.Text = msg
	e.errorText.Refresh()
	e.box.Show()
}

func (e *conditionErrorBox) hide() {
	e.errorText.Text = ""
	e.errorText.Refresh()
	e.box.Hide()
}

func showConditionDetailDialog(
	rule *models.TriggerRule,
	state *OperationsModalState,
	validator ExpressionValidator,
	onUpdated func(*models.TriggerRule),
	parent fyne.Window,
) {
	title := "Edit Condition"
	if rule == nil {
		title = "New Condition"
	}

	labelEntry := buildLabelEntry(rule)
	expressionEntry := buildExpressionEntry(rule)
	errBox := newConditionErrorBox()

	var d dialog.Dialog

	saveBtn := buildConditionSaveBtn(rule, state, labelEntry, expressionEntry, errBox, validator, onUpdated, &d)
	cancelBtn := widget.NewButton("Cancel", func() { d.Hide() })

	content := buildConditionDialogContent(labelEntry, expressionEntry, errBox, saveBtn, cancelBtn)

	d = dialog.NewCustomWithoutButtons(title, content, parent)
	d.Resize(fyne.NewSize(420, 300))
	d.Show()
}

func buildLabelEntry(rule *models.TriggerRule) *widget.Entry {
	e := widget.NewEntry()
	if rule != nil {
		e.SetText(rule.Label)
	}
	return e
}

func buildExpressionEntry(rule *models.TriggerRule) *widget.Entry {
	e := widget.NewMultiLineEntry()
	if rule != nil {
		e.SetText(rule.Expression)
	}
	return e
}

func buildConditionSaveBtn(
	rule *models.TriggerRule,
	state *OperationsModalState,
	labelEntry *widget.Entry,
	expressionEntry *widget.Entry,
	errBox *conditionErrorBox,
	validator ExpressionValidator,
	onUpdated func(*models.TriggerRule),
	d *dialog.Dialog,
) *widget.Button {
	btn := widget.NewButton("Save", func() {
		if err := validator(expressionEntry.Text); err != nil {
			errBox.show(err.Error())
			return
		}
		errBox.hide()
		if rule == nil {
			rule = &models.TriggerRule{
				MaintainLink: false,
				Edited:       true,
			}
		}
		rule.Label = labelEntry.Text
		rule.Expression = expressionEntry.Text
		rule.Edited = true
		state.Differentiator.OnLabelChanged(state.SelectedRules)
		state.Differentiator.OnExpressionChanged(state.SelectedRules)
		state.NotifyChanged()
		onUpdated(rule)
		(*d).Hide()
	})
	btn.Importance = widget.HighImportance
	return btn
}

func buildConditionDialogContent(
	labelEntry *widget.Entry,
	expressionEntry *widget.Entry,
	errBox *conditionErrorBox,
	saveBtn *widget.Button,
	cancelBtn *widget.Button,
) fyne.CanvasObject {
	form := widget.NewForm(
		widget.NewFormItem("Label", labelEntry),
		widget.NewFormItem("Expression", expressionEntry),
	)
	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn, layout.NewSpacer())
	return container.NewBorder(nil, buttons, nil, nil,
		container.NewVBox(form, errBox.box),
	)
}
