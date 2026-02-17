package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type fyneObj struct{ o fyne.CanvasObject }

func (x fyneObj) Native() any   { return x.o }
func (fyneObj) UIObjectMarker() {}

type fyneTab struct{ t *container.TabItem }

func (x fyneTab) Native() any { return x.t }

func wrap(o fyne.CanvasObject) uport.UIObject {
	return fyneObj{o: o}
}

func unwrap(obj uport.UIObject) fyne.CanvasObject {
	return obj.Native().(fyne.CanvasObject)
}
