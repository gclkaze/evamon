package executionlog

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	activeExec *models.TriggerExecution

	isMaximized     bool
	maximizedWindow fyne.Window
	onHide          func()
	onRestore       func()

	detachedTabs     *container.AppTabs
	detachedDiagMap  map[string]*ExecutionDiagramTab
	detachedFilter   *widget.Entry
	detachedOnActive func(*models.TriggerExecution)
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
	p.buildUI()
	return p
}

func (p *ExecutionLogPanel) buildUI() {
	p.toolbar = NewExecutionLogToolbar()
	p.toolbar.OnMaximize = p.onMaximize
	p.toolbar.OnDownload = func() { p.runDownload(p.activeExec) }
	p.toolbar.OnFilterChanged = p.applyMainFilter
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
}

// SetSplitCallbacks injects hide/restore callbacks so maximize can collapse the dashboard split.
func (p *ExecutionLogPanel) SetSplitCallbacks(onHide, onRestore func()) {
	p.onHide = onHide
	p.onRestore = onRestore
}

// SetNameResolver replaces the diagram-ID-to-display-name resolver.
func (p *ExecutionLogPanel) SetNameResolver(fn func(string) string) {
	p.nameResolver = fn
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

func (p *ExecutionLogPanel) applyMainFilter(text, typeFilter string) {
	for _, dt := range p.diagTabs {
		dt.ApplyFilter(text, typeFilter)
	}
}

func (p *ExecutionLogPanel) mainOnActive(exec *models.TriggerExecution) {
	p.activeExec = exec
	if exec != nil {
		p.toolbar.SetDownloadEnabled(true)
	} else {
		p.toolbar.SetDownloadEnabled(false)
	}
}

func (p *ExecutionLogPanel) onMaximize() {
	if p.isMaximized {
		p.maximizedWindow.RequestFocus()
		return
	}
	p.isMaximized = true
	if p.onHide != nil {
		p.onHide()
	}
	w := fyne.CurrentApp().NewWindow("Execution Logs")
	p.maximizedWindow = w
	w.SetContent(p.buildDetachedContent())
	w.Resize(fyne.NewSize(1000, 650))
	w.SetOnClosed(p.onMaximizedClosed)
	w.Show()
}

func (p *ExecutionLogPanel) onMaximizedClosed() {
	p.isMaximized = false
	p.maximizedWindow = nil
	p.detachedTabs = nil
	p.detachedDiagMap = nil
	p.detachedFilter = nil
	p.detachedOnActive = nil
	if p.onRestore != nil {
		p.onRestore()
	}
}

func (p *ExecutionLogPanel) buildDetachedContent() fyne.CanvasObject {
	downloadBtn := widget.NewButton("⬇ Download JSON", nil)
	downloadBtn.Disable()
	p.setupDetachedState(downloadBtn)
	toolbar := p.buildDetachedToolbar(downloadBtn)
	p.detachedTabs = container.NewAppTabs()
	p.detachedDiagMap = make(map[string]*ExecutionDiagramTab)
	return container.NewBorder(toolbar, nil, nil, nil, p.detachedTabs)
}

func (p *ExecutionLogPanel) setupDetachedState(downloadBtn *widget.Button) {
	var activeExec *models.TriggerExecution
	downloadBtn.OnTapped = func() { p.runDownloadTo(activeExec, p.maximizedWindow) }
	p.detachedOnActive = func(exec *models.TriggerExecution) {
		activeExec = exec
		if exec != nil {
			downloadBtn.Enable()
		} else {
			downloadBtn.Disable()
		}
	}
}

func (p *ExecutionLogPanel) buildDetachedToolbar(downloadBtn *widget.Button) fyne.CanvasObject {
	p.detachedFilter = widget.NewEntry()
	p.detachedFilter.SetPlaceHolder("Filter lines...")
	p.detachedFilter.OnChanged = p.applyDetachedFilter
	row1 := container.NewHBox(layout.NewSpacer(), downloadBtn)
	row2 := container.NewBorder(nil, nil, nil, nil, p.detachedFilter)
	return container.NewVBox(row1, row2)
}

func (p *ExecutionLogPanel) applyDetachedFilter(_ string) {
	if p.detachedDiagMap == nil {
		return
	}
	for _, dt := range p.detachedDiagMap {
		dt.ApplyFilter(p.detachedFilter.Text, "All")
	}
}

func (p *ExecutionLogPanel) runDownload(exec *models.TriggerExecution) {
	wins := fyne.CurrentApp().Driver().AllWindows()
	if len(wins) == 0 {
		return
	}
	p.runDownloadTo(exec, wins[0])
}

func (p *ExecutionLogPanel) runDownloadTo(exec *models.TriggerExecution, win fyne.Window) {
	if exec == nil || win == nil {
		return
	}
	d := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil || f == nil {
			return
		}
		defer f.Close()
		data, _ := json.MarshalIndent(exec, "", "  ")
		f.Write(data) //nolint:errcheck
	}, win)
	d.SetFileName(fmt.Sprintf("execution_%s.json", exec.TriggerID))
	d.Show()
}

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
	changed := p.syncTabs(snap, p.tabs, p.diagTabs, p.toolbar.FilterText(), p.mainOnActive)
	if changed {
		p.emptyLabel.Hide()
		p.tabs.Refresh()
	}
	if p.isMaximized && p.detachedTabs != nil {
		if p.syncTabs(snap, p.detachedTabs, p.detachedDiagMap, p.detachedFilter.Text, p.detachedOnActive) {
			p.detachedTabs.Refresh()
		}
	}
}

func (p *ExecutionLogPanel) syncTabs(
	snap map[string]map[string][]*models.TriggerExecution,
	tabs *container.AppTabs,
	diagMap map[string]*ExecutionDiagramTab,
	filterText string,
	onActive func(*models.TriggerExecution),
) bool {
	diagIDs := sortedStringKeys(snap)
	changed := false
	for _, diagID := range diagIDs {
		if _, ok := diagMap[diagID]; !ok {
			dt := NewExecutionDiagramTab(p.renderer, onActive)
			dt.ApplyFilter(filterText, "All")
			diagMap[diagID] = dt
			tabs.Append(container.NewTabItem(truncate(p.nameResolver(diagID), 15), dt.Object()))
			changed = true
		}
		diagMap[diagID].Refresh(snap[diagID])
	}
	return changed
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
