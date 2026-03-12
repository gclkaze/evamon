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

	c        port.Controls
	actions  port.DiagramActions
	registry *port.ChartRegistry
}

func New() *Renderer {
	cr := port.NewChartRegistry()
	return &Renderer{
		a:        app.New(),
		l:        Layout{},
		c:        Controls{},
		actions:  DiagramActionHandler{ChartRegistry:cr},
		registry: cr,
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

func (r *Renderer) ChartRegistry() *port.ChartRegistry {
	return r.registry
}

/*
func (r *Renderer) GetDiagram(id string) *port.ChartRegistry {
	return r.registry.
}*/
