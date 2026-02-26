package fynediagrams

import (
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

	drawer *LineChartDrawer
	root   *fyne.Container
	card   *widget.Card

	// track last size (optional)
	size fyne.Size

	minWidth  float32
	minHeight float32

	chart *fyne.Container
}

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

	// Put chart in a max container so it expands nicely inside the card
	chart := container.NewMax(w.drawer.Object())

	w.card = widget.NewCard(w.title, w.description, chart)
	w.root = container.NewMax(w.card)

	w.size = fyne.NewSize(width, height)
	w.root.Resize(w.size)

	return w
}
func (w *LineChartWidget) CreateRenderer() fyne.WidgetRenderer {
	return &LineChartWidgetRenderer{
		widget: w,
		objects: []fyne.CanvasObject{
			w.drawer.Object(),
		},
	}
}

func (w *LineChartWidget) ToggleItem(it *port.LegendItem) {
	w.drawer.ToggleItem(it)
}
func (w *LineChartWidget) Title() string       { return w.title }
func (w *LineChartWidget) Description() string { return w.description }

func (w *LineChartWidget) Native() any { return w }

func (w *LineChartWidget) Object() fyne.CanvasObject {
	return w.root
}

// Push adds a point to the drawer.
// IMPORTANT: Call this on the UI thread (RunOnMain) if you’re pushing from goroutines.
func (w *LineChartWidget) Push(at time.Time, val any) {
	w.drawer.Push(at, val)
}

// Optional: allow external resizing (if you use WithoutLayout somewhere)
func (w *LineChartWidget) Resize(size fyne.Size) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.size = size
	w.root.Resize(size)
	w.drawer.Resize(size.Width, size.Height)
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

type LineChartWidgetRenderer struct {
	widget  *LineChartWidget
	objects []fyne.CanvasObject
}

func (r *LineChartWidgetRenderer) Layou2t(size fyne.Size) {
	// Resize drawer to fill entire widget
	r.widget.drawer.Resize(size.Width, size.Height)
	r.widget.drawer.Object().Resize(size)
}

func (r *LineChartWidgetRenderer) Layout(size fyne.Size) {
	// size the card/root
	r.widget.root.Resize(size)

	// VERY IMPORTANT: ensure the drawer's root canvas object is also resized,
	// otherwise drawer.root.Size() stays {0,0}.
	r.widget.drawer.Object().Resize(size)

	// and inform the drawer so it can redraw using this size
	r.widget.drawer.Resize(size.Width, size.Height)
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
