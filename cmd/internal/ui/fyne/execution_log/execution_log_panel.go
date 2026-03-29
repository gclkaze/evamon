package executionlog

import (
	"context"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// LogProvider is the read-only interface the panel needs from the backend.
type LogProvider interface {
	Snapshot() map[string]map[string][]*models.TriggerExecution
}

// ExecutionLogPanel is the full log pane: toolbar on top, per-diagram tabs below.
// It implements port.UIObject so it can flow through the abstract layout system.
type ExecutionLogPanel struct {
	provider        LogProvider
	nameResolver    func(string) string
	refreshInterval time.Duration
	renderer        *LogLineRenderer

	root       *fyne.Container
	toolbar    *ExecutionLogToolbar
	tabs       *container.AppTabs
	emptyLabel *widget.Label
	diagTabs   map[string]*ExecutionDiagramTab
}

func NewExecutionLogPanel(provider LogProvider, nameResolver func(string) string, renderer *LogLineRenderer) *ExecutionLogPanel {
	if nameResolver == nil {
		nameResolver = shortID
	}
	p := &ExecutionLogPanel{
		provider:        provider,
		nameResolver:    nameResolver,
		refreshInterval: 2 * time.Second,
		renderer:        renderer,
		diagTabs:        make(map[string]*ExecutionDiagramTab),
	}

	p.toolbar = NewExecutionLogToolbar()
	p.toolbar.OnFilterChanged = func(text, typeFilter string) {
		for _, dt := range p.diagTabs {
			dt.ApplyFilter(text, typeFilter)
		}
	}

	p.tabs = container.NewAppTabs()
	p.emptyLabel = widget.NewLabelWithStyle(
		"No execution logs yet.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	p.root = container.NewBorder(
		p.toolbar.Object(), nil, nil, nil,
		container.NewStack(container.NewCenter(p.emptyLabel), p.tabs),
	)
	return p
}

// StartRefresh starts the background auto-refresh goroutine.
func (p *ExecutionLogPanel) StartRefresh(ctx context.Context) {
	go p.refreshLoop(ctx)
}

// Object returns the root Fyne container for embedding in a VSplit.
func (p *ExecutionLogPanel) Object() *fyne.Container { return p.root }

// port.UIObject implementation
func (p *ExecutionLogPanel) Native() any                  { return p.root }
func (p *ExecutionLogPanel) Title() string                { return "Execution Logs" }
func (p *ExecutionLogPanel) Description() string          { return "" }
func (p *ExecutionLogPanel) ToggleItem(*uport.LegendItem) {}
func (p *ExecutionLogPanel) Refresh()                     { p.root.Refresh() }

func (p *ExecutionLogPanel) refreshLoop(ctx context.Context) {
	tick := time.NewTicker(p.refreshInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			snap := p.provider.Snapshot()
			fyne.Do(func() { p.applySnapshot(snap) })
		}
	}
}

func (p *ExecutionLogPanel) applySnapshot(snap map[string]map[string][]*models.TriggerExecution) {
	diagIDs := sortedStringKeys(snap)
	changed := false
	for _, diagID := range diagIDs {
		if _, ok := p.diagTabs[diagID]; !ok {
			dt := NewExecutionDiagramTab(p.renderer)
			dt.ApplyFilter(p.toolbar.FilterText(), p.toolbar.TypeFilter())
			p.diagTabs[diagID] = dt
			p.tabs.Append(container.NewTabItem(p.nameResolver(diagID), dt.Object()))
			changed = true
		}
		p.diagTabs[diagID].Refresh(snap[diagID])
	}
	if changed {
		p.emptyLabel.Hide()
		p.tabs.Refresh()
	}
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func sortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
