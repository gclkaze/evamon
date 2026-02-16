package fynediagrams

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type BarChartWidget struct {
	widget.BaseWidget
	drawer *BarChartDrawer
}

func NewBarChartWidget(drawer *BarChartDrawer) *BarChartWidget {
	w := &BarChartWidget{drawer: drawer}
	w.ExtendBaseWidget(w)
	return w
}

func (w *BarChartWidget) Push(at time.Time, val any) {
	w.drawer.Push(at, val)
}

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
	return fyne.NewSize(200, 120)
}

func (r *barChartWidgetRenderer) Refresh() {
	r.w.drawer.Redraw()
	canvas.Refresh(r.root)
}

func (r *barChartWidgetRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *barChartWidgetRenderer) Destroy()                     {}
