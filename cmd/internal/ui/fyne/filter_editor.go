package fynerenderer

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynelayout "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	viewproject "github.com/gclkaze/evamon/cmd/internal/viewproject"
)

type FilterEditor struct {
	root fyne.CanvasObject

	titleLabel *widget.Label
	modeRadio  *widget.RadioGroup
	entry      *widget.Entry
	addBtn     *widget.Button
	errorLabel *widget.Label

	rowsBox *fyne.Container
	scroll  *container.Scroll

	saveBtn   *widget.Button
	cancelBtn *widget.Button

	components []viewproject.FilterComponent
	mode       viewproject.FilterMode

	validateFn func(string) error
	saveFn     func(*viewproject.Filter) error
	cancelFn   func()
}

func NewFilterEditor(
	title string,
	initial *viewproject.Filter,
	validateFn func(string) error,
	saveFn func(*viewproject.Filter) error,
	cancelFn func(),
) *FilterEditor {
	fe := &FilterEditor{
		validateFn: validateFn,
		saveFn:     saveFn,
		cancelFn:   cancelFn,
	}

	fe.loadInitial(initial)
	fe.initWidgets(title)
	fe.bindEvents()
	fe.root = fe.buildRoot()
	fe.rebuildRows()

	return fe
}

func (fe *FilterEditor) Object() fyne.CanvasObject {
	return fe.root
}

func (fe *FilterEditor) CurrentFilter() *viewproject.Filter {
	if len(fe.components) == 0 {
		return nil
	}

	return &viewproject.Filter{
		Setup: &viewproject.FilterSetup{
			Mode:       fe.mode,
			Components: fe.cloneComponents(),
		},
	}
}

func (fe *FilterEditor) loadInitial(initial *viewproject.Filter) {
	fe.mode = viewproject.FilterModeAND
	fe.components = nil

	if initial == nil || initial.Setup == nil {
		return
	}

	fe.mode = initial.Setup.Mode
	if fe.mode == "" {
		fe.mode = viewproject.FilterModeAND
	}

	fe.components = append([]viewproject.FilterComponent(nil), initial.Setup.Components...)
}

func (fe *FilterEditor) initWidgets(title string) {
	fe.titleLabel = widget.NewLabel(title)
	fe.titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	fe.modeRadio = widget.NewRadioGroup([]string{"AND", "OR"}, nil)
	fe.modeRadio.Horizontal = true
	fe.setModeSelection()

	fe.entry = widget.NewEntry()
	fe.entry.SetPlaceHolder("Enter filter expression")

	fe.addBtn = widget.NewButton("+", fe.onAdd)
	fe.addBtn.Disable()

	fe.errorLabel = widget.NewLabel("")
	fe.errorLabel.Hide()

	fe.rowsBox = container.NewVBox()
	fe.scroll = container.NewVScroll(fe.rowsBox)
	fe.scroll.SetMinSize(fyne.NewSize(460, 180))

	fe.cancelBtn = widget.NewButton("Cancel", fe.onCancel)
	fe.saveBtn = widget.NewButton("Save", fe.onSave)
}

func (fe *FilterEditor) bindEvents() {
	fe.modeRadio.OnChanged = func(v string) {
		fe.mode = fe.parseMode(v)
	}

	fe.entry.OnChanged = func(s string) {
		if strings.TrimSpace(s) == "" {
			fe.addBtn.Disable()
			return
		}
		fe.addBtn.Enable()
	}
}

func (fe *FilterEditor) buildRoot() fyne.CanvasObject {
	return container.NewBorder(
		fe.buildTop(),
		fe.buildBottom(),
		nil,
		nil,
		fe.buildCenter(),
	)
}

func (fe *FilterEditor) buildTop() fyne.CanvasObject {
	return container.NewVBox(
		fe.titleLabel,
		widget.NewSeparator(),
	)
}

func (fe *FilterEditor) buildBottom() fyne.CanvasObject {
	return container.NewHBox(
		fynelayout.NewSpacer(),
		fe.cancelBtn,
		fe.saveBtn,
	)
}

func (fe *FilterEditor) buildCenter() fyne.CanvasObject {
	return container.NewVBox(
		fe.buildModeRow(),
		fe.buildInputRow(),
		fe.errorLabel,
		fe.scroll,
	)
}

func (fe *FilterEditor) buildModeRow() fyne.CanvasObject {
	return container.NewHBox(
		widget.NewLabel("Combine rules with:"),
		fe.modeRadio,
	)
}

func (fe *FilterEditor) buildInputRow() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		nil,
		fe.addBtn,
		fe.entry,
	)
}

func (fe *FilterEditor) setModeSelection() {
	if fe.mode == viewproject.FilterModeOR {
		fe.modeRadio.SetSelected("OR")
		return
	}
	fe.modeRadio.SetSelected("AND")
}

func (fe *FilterEditor) parseMode(v string) viewproject.FilterMode {
	if strings.EqualFold(strings.TrimSpace(v), "OR") {
		return viewproject.FilterModeOR
	}
	return viewproject.FilterModeAND
}

func (fe *FilterEditor) onAdd() {
	expr := strings.TrimSpace(fe.entry.Text)
	if expr == "" {
		return
	}

	if err := fe.validateExpression(expr); err != nil {
		fe.showError(err.Error())
		return
	}

	fe.components = append(fe.components, viewproject.FilterComponent{
		ID:         "",
		Expression: expr,
		Enabled:    true,
	})

	fe.clearInput()
	fe.hideError()
	fe.rebuildRows()
}

func (fe *FilterEditor) onSave() {
	if err := fe.validateAllExpressions(); err != nil {
		fe.showError("Cannot save: " + err.Error())
		return
	}

	if fe.saveFn == nil {
		return
	}

	if err := fe.saveFn(fe.CurrentFilter()); err != nil {
		fe.showError(err.Error())
	}
}

func (fe *FilterEditor) onCancel() {
	if fe.cancelFn != nil {
		fe.cancelFn()
	}
}

func (fe *FilterEditor) validateExpression(expr string) error {
	if fe.validateFn == nil {
		return nil
	}
	return fe.validateFn(expr)
}

func (fe *FilterEditor) validateAllExpressions() error {
	for _, c := range fe.components {
		if err := fe.validateExpression(c.Expression); err != nil {
			return err
		}
	}
	return nil
}

func (fe *FilterEditor) clearInput() {
	fe.entry.SetText("")
	fe.addBtn.Disable()
}

func (fe *FilterEditor) showError(msg string) {
	fe.errorLabel.SetText(msg)
	fe.errorLabel.Show()
}

func (fe *FilterEditor) hideError() {
	fe.errorLabel.SetText("")
	fe.errorLabel.Hide()
}

func (fe *FilterEditor) removeAt(idx int) {
	if idx < 0 || idx >= len(fe.components) {
		return
	}

	fe.components = append(fe.components[:idx], fe.components[idx+1:]...)
	fe.rebuildRows()
}

func (fe *FilterEditor) rebuildRows() {
	fe.rowsBox.Objects = nil

	if len(fe.components) == 0 {
		fe.rowsBox.Add(widget.NewLabel("No filters added yet."))
		fe.rowsBox.Refresh()
		return
	}

	for i, c := range fe.components {
		fe.rowsBox.Add(fe.buildExpressionRow(i, c))
	}

	fe.rowsBox.Refresh()
}

func (fe *FilterEditor) buildExpressionRow(idx int, c viewproject.FilterComponent) fyne.CanvasObject {
	check := widget.NewCheck("enabled", func(v bool) {
		if idx < 0 || idx >= len(fe.components) {
			return
		}
		fe.components[idx].Enabled = v
	})
	check.SetChecked(c.Enabled)

	label := widget.NewLabel(c.Expression)

	removeBtn := widget.NewButton("-", func() {
		fe.removeAt(idx)
	})

	return container.NewHBox(
		check,
		label,
		fynelayout.NewSpacer(),
		removeBtn,
	)
}

func (fe *FilterEditor) cloneComponents() []viewproject.FilterComponent {
	if len(fe.components) == 0 {
		return nil
	}
	out := make([]viewproject.FilterComponent, len(fe.components))
	copy(out, fe.components)
	return out
}
