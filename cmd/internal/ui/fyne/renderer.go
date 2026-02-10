// ui/fyne/renderer.go
package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Renderer struct {
	a fyne.App
}

func New() *Renderer {
	return &Renderer{a: app.New()}
}

func (r *Renderer) NewExecutionWindow(title string) (port.ExecutionWindow, error) {
	w := r.a.NewWindow(title)
	return &FyneWindow{w: w}, nil
}
