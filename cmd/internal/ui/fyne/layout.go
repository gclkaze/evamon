package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Layout struct{}

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
