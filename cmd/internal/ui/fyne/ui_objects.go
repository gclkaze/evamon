package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type fyneObj struct {
	o           fyne.CanvasObject
	title       string
	description string
}

func (x fyneObj) Native() any   { return x.o }
func (x fyneObj) Title() string { return x.title }
func (x fyneObj) ToggleItem(it *port.LegendItem) {
	//x.ToggleItem(it)
}
func (x fyneObj) GetDataSeries() data.IMultiSeriesData {
	return nil
}
func (x fyneObj) Refresh() {
	x.o.Refresh()
}
func (x fyneObj) Description() string { return x.description }

func (fyneObj) UIObjectMarker() {}

type fyneTab struct {
	t           *container.TabItem
	title       string
	description string
}

func (x fyneTab) Native() any         { return x.t }
func (x fyneTab) Title() string       { return x.title }
func (x fyneTab) Description() string { return x.description }

func wrap(o fyne.CanvasObject) port.UIObject {
	return fyneObj{o: o}
}

func unwrap(obj port.UIObject) fyne.CanvasObject {
	return obj.Native().(fyne.CanvasObject)
}
