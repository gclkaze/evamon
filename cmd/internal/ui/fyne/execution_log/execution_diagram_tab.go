package executionlog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// ExecutionDiagramTab holds per-rule tabs for one diagram.
type ExecutionDiagramTab struct {
	tabs       *container.AppTabs
	ruleTabs   map[string]*ExecutionRuleTab
	emptyLabel *widget.Label
	renderer   *LogLineRenderer
	obj        fyne.CanvasObject
}

func NewExecutionDiagramTab(renderer *LogLineRenderer) *ExecutionDiagramTab {
	d := &ExecutionDiagramTab{
		tabs:     container.NewAppTabs(),
		ruleTabs: make(map[string]*ExecutionRuleTab),
		renderer: renderer,
	}
	d.emptyLabel = widget.NewLabelWithStyle(
		"No trigger rules have fired yet.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	d.obj = container.NewStack(container.NewCenter(d.emptyLabel), d.tabs)
	return d
}

// Object returns the Fyne object for this diagram tab's content area.
func (d *ExecutionDiagramTab) Object() fyne.CanvasObject { return d.obj }

// Refresh updates all rule tabs with the latest execution data.
func (d *ExecutionDiagramTab) Refresh(data map[string][]*models.TriggerExecution) {
	ruleIDs := sortedStringKeys(data)
	changed := false
	for _, ruleID := range ruleIDs {
		if _, ok := d.ruleTabs[ruleID]; !ok {
			rt := NewExecutionRuleTab(d.renderer)
			d.ruleTabs[ruleID] = rt
			d.tabs.Append(container.NewTabItem(shortID(ruleID), rt.Object()))
			changed = true
		}
		d.ruleTabs[ruleID].Refresh(data[ruleID])
	}
	if changed {
		d.emptyLabel.Hide()
		d.tabs.Refresh()
	}
}

// ApplyFilter propagates a filter change to all active rule tabs.
func (d *ExecutionDiagramTab) ApplyFilter(text, typeFilter string) {
	for _, rt := range d.ruleTabs {
		rt.ApplyFilter(text, typeFilter)
	}
}
