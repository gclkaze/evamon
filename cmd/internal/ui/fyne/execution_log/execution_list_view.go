package executionlog

import (
	"fmt"
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// ExecutionListView renders a scrollable list of TriggerExecution summaries,
// newest first. Clicking a row invokes onSelect with the chosen execution.
type ExecutionListView struct {
	box        *fyne.Container
	scroll     *container.Scroll
	executions []*models.TriggerExecution
	onSelect   func(*models.TriggerExecution)
}

func NewExecutionListView(onSelect func(*models.TriggerExecution)) *ExecutionListView {
	v := &ExecutionListView{onSelect: onSelect}
	v.box    = container.NewVBox()
	v.scroll = container.NewVScroll(v.box)
	return v
}

// SetExecutions replaces the displayed executions and rebuilds all rows.
func (v *ExecutionListView) SetExecutions(executions []*models.TriggerExecution) {
	v.executions = executions
	v.box.RemoveAll()
	for i, exec := range executions {
		v.box.Add(v.buildRow(i+1, exec))
	}
	v.scroll.Refresh()
}

// Object returns the Fyne canvas object.
func (v *ExecutionListView) Object() fyne.CanvasObject { return v.scroll }

func (v *ExecutionListView) buildRow(index int, exec *models.TriggerExecution) fyne.CanvasObject {
	indexLabel  := widget.NewLabel(fmt.Sprintf("#%d", index))
	timeLabel   := widget.NewLabel(exec.StartedAt.Format("15:04:05"))
	fileLabel   := widget.NewLabel(filepath.Base(exec.FilePath))
	statusLabel := v.buildStatusLabel(exec)
	row := container.NewHBox(indexLabel, timeLabel, fileLabel, statusLabel)
	return v.wrapClickable(row, exec)
}

func (v *ExecutionListView) buildStatusLabel(exec *models.TriggerExecution) *canvas.Text {
	if exec.FinishedAt == nil {
		return canvas.NewText("…", theme.ForegroundColor())
	}
	if *exec.Success {
		return canvas.NewText("✓", color.NRGBA{R: 0, G: 180, B: 0, A: 255})
	}
	return canvas.NewText("✗", color.NRGBA{R: 200, G: 0, B: 0, A: 255})
}

func (v *ExecutionListView) wrapClickable(content fyne.CanvasObject, exec *models.TriggerExecution) fyne.CanvasObject {
	btn := widget.NewButton("", func() { v.onSelect(exec) })
	return container.NewStack(btn, content)
}
