package fynediagrams

import (
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// LineChartWidget hosts a LineChartDrawer and exposes Push(at,val).
// You can wrap it in your UIObject adapter (wrap/unwrap) like your BoolFill.
type LineChartWidget struct {
	widget.BaseWidget
	mu sync.Mutex

	title       string
	description string

	drawer    *LineChartDrawer
	lastFetch *lastUpdatedLabel
	tooltip   *chartTooltip
	root      *fyne.Container
	card      *widget.Card

	// track last size (optional)
	size fyne.Size

	minWidth  float32
	minHeight float32

	chart *fyne.Container
}

func (w *LineChartWidget) GetDataSeries() data.IMultiSeriesData {
	return w.drawer.GetDataSeries()
}

// MaximizeView returns a widget suitable as a maximize window's content.
// See BarChartWidget.MaximizeView for the full explanation.
// Call ClearMaximizeHook() when the maximize window closes.
func (w *LineChartWidget) MaximizeView(initialW, initialH float32) fyne.CanvasObject {
	w.drawer.Resize(initialW, initialH)
	adapter := newMaximizeAdapter(w.drawer.Object(), func(width, height float32) {
		w.drawer.Resize(width, height)
	})
	w.drawer.refreshHook = func() { adapter.Refresh() }
	return adapter
}

// ClearMaximizeHook removes the refresh hook set by MaximizeView.
func (w *LineChartWidget) ClearMaximizeHook() { w.drawer.refreshHook = nil }

func NewLineChartWidget(drawer *LineChartDrawer, title, description string, width, height float32) *LineChartWidget {
	w := &LineChartWidget{
		title:       strings.TrimSpace(title),
		description: strings.TrimSpace(description),
		minWidth:    width, minHeight: height,
	}
	w.ExtendBaseWidget(w)

	if w.title == "" {
		w.title = "Line chart"
	}

	w.drawer = drawer
	w.lastFetch = newLastUpdatedLabel()
	w.tooltip = newChartTooltip()

	// Put chart in a max container so it expands nicely inside the card
	chart := container.NewMax(w.drawer.Object())

	w.card = widget.NewCard(w.title, w.description, chart)
	w.root = container.NewMax(w.card)

	w.size = fyne.NewSize(width, height)
	w.root.Resize(w.size)

	return w
}

func (w *LineChartWidget) Refresh() {
	w.BaseWidget.Refresh()
}

func (w *LineChartWidget) CreateRenderer() fyne.WidgetRenderer {
	return &LineChartWidgetRenderer{
		widget: w,
		objects: []fyne.CanvasObject{
			w.drawer.Object(),
			w.tooltip.Object(),
		},
	}
}

func (w *LineChartWidget) ToggleItem(it *port.LegendItem) {
	w.drawer.ToggleItem(it)
}

func (w *LineChartWidget) ZoomIn() {
	w.drawer.ZoomIn()
}

func (w *LineChartWidget) ZoomOut() {
	w.drawer.ZoomOut()
}

func (w *LineChartWidget) Title() string       { return w.title }
func (w *LineChartWidget) Description() string { return w.description }

func (w *LineChartWidget) Native() any { return w }

func (w *LineChartWidget) Object() fyne.CanvasObject {
	return w.root
}

// Push adds a point to the drawer and updates the last-fetch label.
// IMPORTANT: Call this on the UI thread (RunOnMain) if you're pushing from goroutines.
func (w *LineChartWidget) Push(at time.Time, val any) {
	w.drawer.Push(at, val)
	w.lastFetch.update(at)
}

func (w *LineChartWidget) LastUpdatedLabel() uport.UIObject {
	return w.lastFetch
}

func (w *LineChartWidget) Resize(size fyne.Size) {
	// IMPORTANT: must call BaseWidget.Resize so baseObject.size is updated.
	// Fyne uses BaseWidget.Size() for hit-testing (hover, tap, etc.).
	// Without this, the widget appears zero-sized to the event system and
	// never receives desktop.Hoverable (or any pointer) events.
	// BaseWidget.Resize also calls Renderer.Layout, which handles
	// w.root, drawer, and tooltip resizing.
	w.BaseWidget.Resize(size)
}

// Optional: if you ever want to change title at runtime
func (w *LineChartWidget) SetTitle(title string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.title = strings.TrimSpace(title)
	w.card.SetTitle(w.title)
}

func (w *LineChartWidget) SetDescription(desc string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.description = strings.TrimSpace(desc)
	w.card.SetSubTitle(w.description)
}

// desktop.Hoverable implementation — shows a per-point tooltip on mouse hover.

func (w *LineChartWidget) MouseIn(_ *desktop.MouseEvent) {}

func (w *LineChartWidget) MouseOut() {
	w.tooltip.Hide()
}

func (w *LineChartWidget) MouseMoved(e *desktop.MouseEvent) {
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

type LineChartWidgetRenderer struct {
	widget  *LineChartWidget
	objects []fyne.CanvasObject
}

func (r *LineChartWidgetRenderer) Layout(size fyne.Size) {
	r.widget.size = size

	// size the card/root
	r.widget.root.Resize(size)

	// VERY IMPORTANT: ensure the drawer's root canvas object is also resized,
	// otherwise drawer.root.Size() stays {0,0}.
	r.widget.drawer.Object().Resize(size)

	// and inform the drawer so it can redraw using this size
	r.widget.drawer.Resize(size.Width, size.Height)

	// give the tooltip overlay the full widget area so it can position freely
	r.widget.tooltip.Object().Resize(size)
}

func (r *LineChartWidgetRenderer) MinSize() fyne.Size {
	//300,150
	return fyne.NewSize(r.widget.minWidth, r.widget.minHeight)
}

func (r *LineChartWidgetRenderer) Refresh() {
	r.widget.drawer.Object().Refresh()
}

func (r *LineChartWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *LineChartWidgetRenderer) Destroy() {}

var _ dport.DiagramWidget = (*LineChartWidget)(nil)
var _ uport.UIObject = (*LineChartWidget)(nil)
var _ desktop.Hoverable = (*LineChartWidget)(nil)
