package executionlog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ExecutionLogToolbar is the two-row control bar at the top of the log panel.
// Row 1: action buttons. Row 2: text filter + type selector.
type ExecutionLogToolbar struct {
	obj             *fyne.Container
	filterEntry     *widget.Entry
	typeSelect      *widget.Select
	OnFilterChanged func(text, typeFilter string)
}

// NewExecutionLogToolbar builds the toolbar. Connect OnFilterChanged after
// construction to receive filter-change events.
func NewExecutionLogToolbar() *ExecutionLogToolbar {
	t := &ExecutionLogToolbar{}

	maximizeBtn := widget.NewButton("⛶ Maximize", func() {})
	detachBtn   := widget.NewButton("⧉ Detach", func() {})
	downloadBtn := widget.NewButton("⬇ Download", func() {})
	row1 := container.NewHBox(maximizeBtn, detachBtn, layout.NewSpacer(), downloadBtn)

	t.filterEntry = widget.NewEntry()
	t.filterEntry.SetPlaceHolder("Filter lines...")

	t.typeSelect = widget.NewSelect([]string{"All", "Operation", "Label", "Program"}, nil)
	t.typeSelect.SetSelected("All")

	notify := func() {
		if t.OnFilterChanged != nil {
			t.OnFilterChanged(t.filterEntry.Text, t.typeSelect.Selected)
		}
	}
	t.filterEntry.OnChanged = func(_ string) { notify() }
	t.typeSelect.OnChanged  = func(_ string) { notify() }

	row2 := container.NewBorder(nil, nil, nil, t.typeSelect, t.filterEntry)
	t.obj = container.NewVBox(row1, row2)
	return t
}

// Object returns the Fyne canvas object for embedding in a container.
func (t *ExecutionLogToolbar) Object() fyne.CanvasObject { return t.obj }

// FilterText returns the current text filter value.
func (t *ExecutionLogToolbar) FilterText() string { return t.filterEntry.Text }

// TypeFilter returns the currently selected type filter value.
func (t *ExecutionLogToolbar) TypeFilter() string { return t.typeSelect.Selected }
