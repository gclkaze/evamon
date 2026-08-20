package executionlog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ExecutionLogToolbar is the two-row control bar at the top of the log panel.
// Row 1: Maximize (left) + Download JSON (right).
// Row 2: text filter (full width). Type filter is hardcoded to "All".
type ExecutionLogToolbar struct {
	obj             *fyne.Container
	filterEntry     *widget.Entry
	downloadBtn     *widget.Button
	OnFilterChanged func(text, typeFilter string)
	OnMaximize      func()
	OnDownload      func()
}

// NewExecutionLogToolbar builds the toolbar with the download button initially disabled.
func NewExecutionLogToolbar() *ExecutionLogToolbar {
	t := &ExecutionLogToolbar{}

	maximizeBtn := widget.NewButton("⛶ Maximize", func() {
		if t.OnMaximize != nil {
			t.OnMaximize()
		}
	})
	t.downloadBtn = widget.NewButton("⬇ Download JSON", func() {
		if t.OnDownload != nil {
			t.OnDownload()
		}
	})
	t.downloadBtn.Disable()

	row1 := container.NewHBox(maximizeBtn, layout.NewSpacer(), t.downloadBtn)

	t.filterEntry = widget.NewEntry()
	t.filterEntry.SetPlaceHolder("Filter lines...")
	t.filterEntry.OnChanged = func(_ string) { t.notify() }

	row2 := container.NewBorder(nil, nil, nil, nil, t.filterEntry)
	t.obj = container.NewVBox(row1, row2)
	return t
}

func (t *ExecutionLogToolbar) notify() {
	if t.OnFilterChanged != nil {
		t.OnFilterChanged(t.filterEntry.Text, "All")
	}
}

// SetDownloadEnabled enables or disables the download button.
func (t *ExecutionLogToolbar) SetDownloadEnabled(enabled bool) {
	if enabled {
		t.downloadBtn.Enable()
	} else {
		t.downloadBtn.Disable()
	}
}

// Object returns the Fyne canvas object for embedding in a container.
func (t *ExecutionLogToolbar) Object() fyne.CanvasObject { return t.obj }

// FilterText returns the current text filter value.
func (t *ExecutionLogToolbar) FilterText() string { return t.filterEntry.Text }

// TypeFilter always returns "All" — the type dropdown is hidden.
func (t *ExecutionLogToolbar) TypeFilter() string { return "All" }
