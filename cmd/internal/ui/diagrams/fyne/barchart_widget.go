package fynediagrams

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type BarChartWidget struct {
	widget.BaseWidget
	drawer      *BarChartDrawer
	lastFetch   *lastUpdatedLabel
	tooltip     *chartTooltip
	title       string
	description string

	size fyne.Size

	minWidth  float32
	minHeight float32
}

// MaximizeView returns a widget suitable as a maximize window's content.
//   - Pre-sizes d.root so redrawWithAxis() doesn't exit early on the first
//     data push (the readRootSize guard fires when size is {0,0}).
//   - Installs a refreshHook on the drawer so every d.refreshRoot() call
//     goes through adapter.Refresh() → CanvasForObject(adapter) (reliable,
//     adapter IS window content) rather than CanvasForObject(d.root) which
//     can be stale or nil when d.root is inside a widget renderer.
//
// Call ClearMaximizeHook() when the maximize window closes.
func (w *BarChartWidget) MaximizeView(initialW, initialH float32) fyne.CanvasObject {
	w.drawer.root.Resize(fyne.NewSize(initialW, initialH))
	adapter := newMaximizeAdapter(w.drawer.Root(), func(width, height float32) {
		w.drawer.root.Resize(fyne.NewSize(width, height))
		w.drawer.Redraw()
	})
	w.drawer.refreshHook = func() { adapter.Refresh() }
	return adapter
}

// ClearMaximizeHook removes the refresh hook set by MaximizeView.
// Call this when the maximize window is closed so normal rendering resumes.
func (w *BarChartWidget) ClearMaximizeHook() { w.drawer.refreshHook = nil }

func NewBarChartWidget(drawer *BarChartDrawer, title string, description string, width, height float32) *BarChartWidget {
	w := &BarChartWidget{
		drawer:      drawer,
		lastFetch:   newLastUpdatedLabel(),
		tooltip:     newChartTooltip(),
		title:       title,
		description: description,
		minWidth:    width,
		minHeight:   height,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *BarChartWidget) Push(at time.Time, val any) {
	w.drawer.Push(at, val)
	w.lastFetch.update(at)
}

func (w *BarChartWidget) LastUpdatedLabel() port.UIObject { return w.lastFetch }
func (w *BarChartWidget) Title() string                   { return w.title }
func (w *BarChartWidget) Description() string             { return w.description }
func (w *BarChartWidget) ToggleItem(it *port.LegendItem) {
	w.drawer.ToggleItem(it)
}

func (w *BarChartWidget) ZoomIn() {
	w.drawer.ZoomIn()
}

func (w *BarChartWidget) ZoomOut() {
	w.drawer.ZoomOut()
}

func (w *BarChartWidget) Refresh() {
	w.BaseWidget.Refresh()
}

func (w *BarChartWidget) GetDataSeries() data.IMultiSeriesData {
	return w.drawer.GetDataSeries()
}

func (w *BarChartWidget) SetTriggerSender(fn data.TriggerSendFunc) {
	if ds := w.GetDataSeries(); ds != nil {
		ds.SetTriggerSender(fn)
	}
}

func (w *BarChartWidget) Native() any { return w }

func (w *BarChartWidget) CreateRenderer() fyne.WidgetRenderer {
	root := w.drawer.Root()
	return &barChartWidgetRenderer{
		w:    w,
		root: root,
		objs: []fyne.CanvasObject{root, w.tooltip.Object()},
	}
}

// desktop.Hoverable implementation.

func (w *BarChartWidget) MouseIn(_ *desktop.MouseEvent) {}

func (w *BarChartWidget) MouseOut() {
	w.tooltip.Hide()
}

func (w *BarChartWidget) MouseMoved(e *desktop.MouseEvent) {
	idx, times, values, found := w.drawer.HitTestX(e.Position.X, w.size.Width)
	if !found || idx >= len(times) {
		w.tooltip.Hide()
		return
	}
	vars := w.drawer.Variables()
	shown := w.drawer.ShownIndices()
	seriesVals := make([]int, len(vars))
	for j := range vars {
		if j < len(values) && idx < len(values[j]) {
			seriesVals[j] = values[j][idx]
		}
	}
	w.tooltip.Update(e.Position, w.size, times[idx], vars, seriesVals, shown)
}

type barChartWidgetRenderer struct {
	w    *BarChartWidget
	root fyne.CanvasObject
	objs []fyne.CanvasObject
}

func (r *barChartWidgetRenderer) Layout(size fyne.Size) {
	r.w.size = size
	// Propagate size to the drawer root container.
	r.root.Resize(size)
	r.w.drawer.root.Resize(size)
	// Redraw using the new size (drawer reads d.root.Size()).
	r.w.drawer.Redraw()
	// Give the tooltip overlay the full widget area.
	r.w.tooltip.Object().Resize(size)
}

func (r *barChartWidgetRenderer) MinSize() fyne.Size {
	//400, 280
	return fyne.NewSize(r.w.minWidth, r.w.minHeight)
}

func (r *barChartWidgetRenderer) Refresh() {
	r.w.drawer.Redraw()
	canvas.Refresh(r.root)
}

func (r *barChartWidgetRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *barChartWidgetRenderer) Destroy()                     {}

var _ desktop.Hoverable = (*BarChartWidget)(nil)
