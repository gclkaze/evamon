package executionlog

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// ExecutionDetailView shows the rendered log lines of a single TriggerExecution.
// Filtering is driven externally via ApplyFilter — the toolbar owns the filter UI.
type ExecutionDetailView struct {
	obj         fyne.CanvasObject
	linesBox    *fyne.Container
	scroll      *container.Scroll
	renderer    *LogLineRenderer
	currentExec *models.TriggerExecution
	filterText  string
	typeFilter  string
}

func NewExecutionDetailView(onBack func(), renderer *LogLineRenderer) *ExecutionDetailView {
	v := &ExecutionDetailView{renderer: renderer, typeFilter: "All"}
	backBtn    := widget.NewButton("← Back", onBack)
	header     := container.NewHBox(backBtn)
	v.linesBox  = container.NewVBox()
	v.scroll    = container.NewVScroll(v.linesBox)
	v.obj       = container.NewBorder(header, nil, nil, nil, v.scroll)
	return v
}

// SetExecution loads a new execution and rebuilds the displayed lines.
func (v *ExecutionDetailView) SetExecution(exec *models.TriggerExecution) {
	v.currentExec = exec
	v.rebuildLines()
}

// ApplyFilter updates the active filter and re-renders the current execution.
// Called by ExecutionRuleTab when the toolbar filter changes.
func (v *ExecutionDetailView) ApplyFilter(text, typeFilter string) {
	v.filterText = text
	v.typeFilter = typeFilter
	v.rebuildLines()
}

// Object returns the Fyne canvas object for embedding in a container.
func (v *ExecutionDetailView) Object() fyne.CanvasObject { return v.obj }

func (v *ExecutionDetailView) rebuildLines() {
	v.linesBox.RemoveAll()
	if v.currentExec == nil {
		v.scroll.Refresh()
		return
	}
	for _, entry := range v.currentExec.Lines {
		if v.linePassesFilter(entry) {
			v.linesBox.Add(v.renderer.Render(entry))
		}
	}
	v.scroll.Refresh()
}

func (v *ExecutionDetailView) linePassesFilter(entry models.ExecutionLogEntry) bool {
	if v.filterText != "" && !strings.Contains(strings.ToLower(entry.ParsedMsg), strings.ToLower(v.filterText)) {
		return false
	}
	if v.typeFilter == "" || v.typeFilter == "All" {
		return true
	}
	pl := ParseLogLine(entry.ParsedMsg)
	return pl.IsStructured && string(pl.OperationType) == v.typeFilter
}
