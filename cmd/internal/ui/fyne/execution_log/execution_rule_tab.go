package executionlog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// ExecutionRuleTab manages list ↔ detail navigation for one rule's executions.
// Clicking an execution row enters the detail view; the back button returns to the list.
type ExecutionRuleTab struct {
	stack      *fyne.Container
	listView   *ExecutionListView
	detailView *ExecutionDetailView
}

func NewExecutionRuleTab(renderer *LogLineRenderer) *ExecutionRuleTab {
	rt := &ExecutionRuleTab{}

	rt.detailView = NewExecutionDetailView(func() { rt.showList() }, renderer)

	rt.listView = NewExecutionListView(func(exec *models.TriggerExecution) {
		rt.detailView.SetExecution(exec)
		rt.showDetail()
	})

	rt.stack = container.NewStack(rt.listView.Object())
	return rt
}

// Object returns the Fyne canvas object to place inside a tab.
func (rt *ExecutionRuleTab) Object() fyne.CanvasObject { return rt.stack }

// Refresh updates the execution list (newest first).
func (rt *ExecutionRuleTab) Refresh(executions []*models.TriggerExecution) {
	rt.listView.SetExecutions(reverseExecs(executions))
}

// ApplyFilter propagates a filter change to the detail view.
func (rt *ExecutionRuleTab) ApplyFilter(text, typeFilter string) {
	rt.detailView.ApplyFilter(text, typeFilter)
}

func (rt *ExecutionRuleTab) showList() {
	rt.stack.RemoveAll()
	rt.stack.Add(rt.listView.Object())
	rt.stack.Refresh()
}

func (rt *ExecutionRuleTab) showDetail() {
	rt.stack.RemoveAll()
	rt.stack.Add(rt.detailView.Object())
	rt.stack.Refresh()
}

func reverseExecs(in []*models.TriggerExecution) []*models.TriggerExecution {
	n := len(in)
	out := make([]*models.TriggerExecution, n)
	for i, v := range in {
		out[n-1-i] = v
	}
	return out
}
