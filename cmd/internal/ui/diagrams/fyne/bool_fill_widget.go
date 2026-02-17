package fynediagrams

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type BoolFillWidget struct {
	widget.BaseWidget
	drawer *BoolFillDrawer
}

func NewBoolFillWidget(drawer *BoolFillDrawer) *BoolFillWidget {
	w := &BoolFillWidget{drawer: drawer}
	w.ExtendBaseWidget(w)
	return w
}
func (w *BoolFillWidget) Native() any { return w }
func (w *BoolFillWidget) CreateRenderer() fyne.WidgetRenderer {
	root := w.drawer.Root()
	return &boolFillWidgetRenderer{w: w, root: root, objs: []fyne.CanvasObject{root}}
}

type boolFillWidgetRenderer struct {
	w    *BoolFillWidget
	root fyne.CanvasObject
	objs []fyne.CanvasObject
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
	return fyne.NewSize(200, 120)
}

func (r *boolFillWidgetRenderer) Refresh() {
	r.w.drawer.Redraw()
	canvas.Refresh(r.root)
}

func (r *boolFillWidgetRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *boolFillWidgetRenderer) Destroy()                     {}

var _ dport.DiagramWidget = (*BoolFillWidget)(nil)
var _ uport.UIObject = (*BoolFillWidget)(nil)
