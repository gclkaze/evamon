// ui/fyne/renderer.go
package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Renderer struct {
	a fyne.App
	l port.Layout
}

func New() *Renderer {
	return &Renderer{
		a: app.New(),
		l: Layout{},
	}
}

func (r *Renderer) NewExecutionWindow(title string) (port.ExecutionWindow, error) {
	w := r.a.NewWindow(title)
	return NewFyneWindow(w), nil
}
func (r *Renderer) Layout() port.Layout { return r.l }
func (r *Renderer) Run() {
	r.a.Run()
}
