package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Layout struct{}

func (Layout) Label(text string) uport.UIObject {
	return wrap(widget.NewLabel(text))
}

func (Layout) VScroll(content port.UIObject) port.UIObject {
	return wrap(container.NewVScroll(unwrap(content)))
}

func (Layout) Title(text string) uport.UIObject {
	return wrap(widget.NewLabelWithStyle(
		text,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	))
}

func (Layout) Separator() uport.UIObject {
	return wrap(widget.NewSeparator())
}

func (Layout) Card(title, subtitle string, content uport.UIObject) uport.UIObject {
	return wrap(widget.NewCard(title, subtitle, unwrap(content)))
}

func (Layout) VBox(children ...port.UIObject) port.UIObject {
	objs := make([]fyne.CanvasObject, 0, len(children))
	for _, c := range children {
		objs = append(objs, unwrap(c))
	}

	return wrap(container.NewVBox(objs...))
}

func (Layout) Max(children ...port.UIObject) port.UIObject {
	objs := make([]fyne.CanvasObject, 0, len(children))
	for _, c := range children {
		objs = append(objs, unwrap(c))
	}
	return wrap(container.NewMax(objs...))
}

func (Layout) GridCols(cols int, children ...port.UIObject) port.UIObject {
	objs := make([]fyne.CanvasObject, 0, len(children))
	for _, c := range children {
		objs = append(objs, unwrap(c))
	}
	return wrap(container.NewGridWithColumns(cols, objs...))
}

func (Layout) Padded(child port.UIObject) port.UIObject {
	return wrap(container.NewPadded(unwrap(child)))
}

func (Layout) Tab(title string, content port.UIObject) port.TabItem {
	return fyneTab{t: container.NewTabItem(title, unwrap(content))}
}

func (Layout) Tabs(items ...port.TabItem) port.UIObject {
	tabs := container.NewAppTabs()
	for _, it := range items {
		ti := it.Native().(*container.TabItem)
		tabs.Append(ti)
	}
	return wrap(tabs)
}
