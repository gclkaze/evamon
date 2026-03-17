package fynediagrams

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// maximizeAdapter is a lightweight widget used as the content of the maximize
// window. Its only job is to forward Layout(size) into the drawer's full resize
// logic so the chart redraws correctly whenever the maximize window is resized.
//
// By having d.root in the renderer's Objects(), the maximize window's canvas
// correctly tracks d.root, so CanvasForObject(d.root) resolves to the right
// canvas and real-time data refreshes reach the maximize window.
type maximizeAdapter struct {
	widget.BaseWidget
	inner    fyne.CanvasObject     // the drawer's root container (d.root)
	resizeFn func(w, h float32)    // full resize logic from the chart widget
}

func newMaximizeAdapter(inner fyne.CanvasObject, resizeFn func(w, h float32)) *maximizeAdapter {
	a := &maximizeAdapter{inner: inner, resizeFn: resizeFn}
	a.ExtendBaseWidget(a)
	return a
}

func (a *maximizeAdapter) CreateRenderer() fyne.WidgetRenderer {
	return &maximizeAdapterRenderer{a: a}
}

type maximizeAdapterRenderer struct {
	a *maximizeAdapter
}

// Layout is called by the window whenever its size changes.
// It mirrors the chart widget's renderer Layout so the drawer sees the correct size.
func (r *maximizeAdapterRenderer) Layout(size fyne.Size) {
	r.a.resizeFn(size.Width, size.Height)
}

func (r *maximizeAdapterRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

func (r *maximizeAdapterRenderer) Refresh() {
	// Use the adapter's own canvas (reliably tracked as window content) to
	// refresh d.root, bypassing CanvasForObject(d.root) which can be stale
	// or nil when d.root lives inside a widget renderer rather than being
	// a direct canvas content object.
	c := fyne.CurrentApp().Driver().CanvasForObject(r.a)
	if c != nil {
		c.Refresh(r.a.inner)
		return
	}
	r.a.inner.Refresh()
}

func (r *maximizeAdapterRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.a.inner}
}

func (r *maximizeAdapterRenderer) Destroy() {}
