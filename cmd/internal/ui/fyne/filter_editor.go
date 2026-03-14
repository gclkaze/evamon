package fynerenderer

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynelayout "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
	dia "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type FilterEditor struct {
	root fyne.CanvasObject

	titleLabel *widget.Label
	modeRadio  *widget.RadioGroup
	entry      *widget.Entry
	addBtn     *widget.Button

	errorText *canvas.Text
	errorIcon *widget.Icon
	errorBox  *fyne.Container

	filterEnabled bool

	rowsBox *fyne.Container
	scroll  *container.Scroll

	saveBtn   *widget.Button
	cancelBtn *widget.Button

	components []models.FilterComponent
	mode       models.FilterMode

	diagram dia.IDiagram

	validateFn func(string) error
	saveFn     func(*models.Filter, dia.IDiagram) error
	cancelFn   func()
}

func NewFilterEditor(
	title string,
	initial *models.Filter,
	diagram dia.IDiagram,
	validateFn func(string) error,
	saveFn func(*models.Filter, dia.IDiagram) error,
	cancelFn func(),
) *FilterEditor {
	fe := &FilterEditor{
		validateFn: validateFn,
		saveFn:     saveFn,
		cancelFn:   cancelFn,
		diagram:    diagram,
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

func (fe *FilterEditor) CurrentFilter() *models.Filter {
	if len(fe.components) == 0 {
		return nil
	}

	return &models.Filter{
		Enabled: fe.filterEnabled,
		Setup: &models.FilterSetup{
			Mode:       fe.mode,
			Components: fe.cloneComponents(),
		},
	}
}

func (fe *FilterEditor) loadInitial(initial *models.Filter) {
	fe.mode = models.FilterModeAND
	fe.components = nil
	fe.filterEnabled = false

	if initial == nil || initial.Setup == nil {
		return
	}
	fe.filterEnabled = initial.Enabled
	fe.mode = initial.Setup.Mode
	if fe.mode == "" {
		fe.mode = models.FilterModeAND
	}
	fe.components = append([]models.FilterComponent(nil), initial.Setup.Components...)
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

	fe.errorText = canvas.NewText("", color.NRGBA{R: 211, G: 47, B: 47, A: 255})
	fe.errorText.TextSize = 12

	fe.errorIcon = widget.NewIcon(theme.WarningIcon())

	fe.errorBox = container.NewHBox(
		fe.errorIcon,
		fe.errorText,
	)

	fe.errorBox.Hide()

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
		fe.errorBox,
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
	if fe.mode == models.FilterModeOR {
		fe.modeRadio.SetSelected("OR")
		return
	}
	fe.modeRadio.SetSelected("AND")
}

func (fe *FilterEditor) parseMode(v string) models.FilterMode {
	if strings.EqualFold(strings.TrimSpace(v), "OR") {
		return models.FilterModeOR
	}
	return models.FilterModeAND
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

	fe.components = append(fe.components, models.NewFilterComponent(expr))

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

	if err := fe.saveFn(fe.CurrentFilter(), fe.diagram); err != nil {
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
	fe.errorText.Text = msg
	fe.errorText.Refresh()
	fe.errorBox.Show()
}
func (fe *FilterEditor) hideError() {
	fe.errorText.Text = ""
	fe.errorText.Refresh()
	fe.errorBox.Hide()
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

/*
	func (fe *FilterEditor) buildExpressionRow(idx int, c models.FilterComponent) fyne.CanvasObject {
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
*/
func (fe *FilterEditor) buildExpressionRow(idx int, c models.FilterComponent) fyne.CanvasObject {
	check := widget.NewCheck("", func(v bool) {
		if idx < 0 || idx >= len(fe.components) {
			return
		}
		fe.components[idx].Enabled = v
	})
	check.SetChecked(c.Enabled)

	// use label if set, otherwise fall back to expression
	label := widget.NewLabel(models.GetFilterLabel(c.Label, c.Expression))

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
func (fe *FilterEditor) cloneComponents() []models.FilterComponent {
	if len(fe.components) == 0 {
		return nil
	}
	out := make([]models.FilterComponent, len(fe.components))
	copy(out, fe.components)
	return out
}
