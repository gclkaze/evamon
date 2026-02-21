package fynediagrams

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type BarChartWidget struct {
	widget.BaseWidget
	drawer      *BarChartDrawer
	title       string
	description string

	minWidth  float32
	minHeight float32
}

func NewBarChartWidget(drawer *BarChartDrawer, title string, description string, width, height float32) *BarChartWidget {
	w := &BarChartWidget{drawer: drawer, title: title, description: description, minWidth: width, minHeight: height}
	w.ExtendBaseWidget(w)
	return w
}

func (w *BarChartWidget) Push(at time.Time, val any) {
	w.drawer.Push(at, val)
}
func (w *BarChartWidget) Title() string       { return w.title }
func (w *BarChartWidget) Description() string { return w.description }

func (w *BarChartWidget) Native() any { return w }
func (w *BarChartWidget) CreateRenderer() fyne.WidgetRenderer {
	root := w.drawer.Root()
	return &barChartWidgetRenderer{
		w:    w,
		root: root,
		objs: []fyne.CanvasObject{root},
	}
}

type barChartWidgetRenderer struct {
	w    *BarChartWidget
	root fyne.CanvasObject
	objs []fyne.CanvasObject
}

func (r *barChartWidgetRenderer) Layout(size fyne.Size) {
	// Propagate size to the drawer root container
	r.root.Resize(size)

	// Redraw using the new size (drawer reads d.root.Size()).
	r.w.drawer.Redraw()
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
