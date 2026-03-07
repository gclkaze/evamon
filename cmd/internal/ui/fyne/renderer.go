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

	c       port.Controls
	actions port.DiagramActions
}

func New() *Renderer {
	return &Renderer{
		a:       app.New(),
		l:       Layout{},
		c:       Controls{},
		actions: DiagramActionHandler{},
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
func (r *Renderer) Controls() port.Controls {
	return r.c
}

func (r *Renderer) Actions() port.DiagramActions {
	return r.actions
}
