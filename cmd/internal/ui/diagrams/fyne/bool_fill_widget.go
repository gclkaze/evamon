package fynediagrams

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type BoolFillWidget struct {
	widget.BaseWidget
	drawer      *BoolFillDrawer
	title       string
	description string

	minWidth  float32
	minHeight float32
}

func NewBoolFillWidget(drawer *BoolFillDrawer, title string, description string, width, height float32) *BoolFillWidget {
	w := &BoolFillWidget{drawer: drawer, title: title, description: description, minWidth: width, minHeight: height}
	w.ExtendBaseWidget(w)
	return w
}
func (w *BoolFillWidget) Native() any         { return w }
func (w *BoolFillWidget) Title() string       { return w.title }
func (w *BoolFillWidget) Description() string { return w.description }
func (x *BoolFillWidget) ToggleItem(it *port.LegendItem) {
	//x.ToggleItem(it)
	fmt.Printf("BoolFillWidget toggle item")
	fmt.Print(it)
}
func (w *BoolFillWidget) CreateRenderer() fyne.WidgetRenderer {
	root := w.drawer.Root()
	return &boolFillWidgetRenderer{w: w, root: root, objs: []fyne.CanvasObject{root}}
}

type boolFillWidgetRenderer struct {
	w    *BoolFillWidget
	root fyne.CanvasObject
	objs []fyne.CanvasObject
}

func (w *BoolFillWidget) Refresh() {
	//	w.drawer.Refresh()
}

func (w *BoolFillWidget) GetDataSeries() data.IMultiSeriesData {
	return w.drawer.GetDataSeries()
}

func (w *BoolFillWidget) Push(at time.Time, val any) {
	w.drawer.Push(at, val)
}

func (r *boolFillWidgetRenderer) Layout(size fyne.Size) {
	// Make sure the drawer root matches the allocated size
	r.root.Resize(size)

	// Trigger a redraw that uses the new size
	// (whatever your bool drawer uses: redraw(), redrawWithAxis(), etc.)
	r.w.drawer.Redraw() // rename to your method
}

func (r *boolFillWidgetRenderer) MinSize() fyne.Size {
	//200, 120
	return fyne.NewSize(r.w.minWidth, r.w.minHeight)
}

func (r *boolFillWidgetRenderer) Refresh() {
	r.w.drawer.Redraw()
	canvas.Refresh(r.root)
}

func (r *boolFillWidgetRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *boolFillWidgetRenderer) Destroy()                     {}

var _ dport.DiagramWidget = (*BoolFillWidget)(nil)
var _ uport.UIObject = (*BoolFillWidget)(nil)
