// ===============================
// FILE: cmd/internal/ui/fynerenderer/layout.go
// ===============================
package fynerenderer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynelayout "fyne.io/fyne/v2/layout"
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

func (Layout) ReplaceHBoxContent(hbox port.UIObject, children ...port.UIObject) {
	c := unwrap(hbox).(*fyne.Container)
	objs := make([]fyne.CanvasObject, 0, len(children))
	for _, child := range children {
		objs = append(objs, unwrap(child))
	}
	c.Objects = objs
	c.Refresh()
}

func (Layout) HScroll(content port.UIObject) port.UIObject {
	scroll := container.NewHScroll(unwrap(content))
	scroll.SetMinSize(fyne.NewSize(0, 50)) // fixed height, free width
	return wrap(scroll)
}

func (Layout) Tabs(items ...port.TabItem) port.UIObject {
	tabs := container.NewAppTabs()
	for _, it := range items {
		ti := it.Native().(*container.TabItem)
		tabs.Append(ti)
	}
	return wrap(tabs)
}

// ----------------------
// NEW: HBox + Spacer
// ----------------------

func (Layout) HBox(children ...port.UIObject) port.UIObject {
	objs := make([]fyne.CanvasObject, 0, len(children))
	for _, c := range children {
		objs = append(objs, unwrap(c))
	}
	return wrap(container.NewHBox(objs...))
}

func (Layout) Spacer() port.UIObject {
	return wrap(fynelayout.NewSpacer())
}

// ----------------------
// NEW: DiagramLegend
// ----------------------

func (Layout) DiagramLegend(items []port.LegendItem, onClick func(*port.LegendItem)) port.UIObject {
	obj := newDiagramLegend(items, 6, onClick) // 6 items per row (tweak)
	return wrap(obj)
}

func (Layout) SetLegendAction(legend port.UIObject, onClick func(*port.LegendItem)) {
	unwrap(legend)
}
func (Layout) Border(top, bottom, left, right, center port.UIObject) port.UIObject {
	var topObj, bottomObj, leftObj, rightObj, centerObj fyne.CanvasObject

	if top != nil {
		topObj = unwrap(top)
	}
	if bottom != nil {
		bottomObj = unwrap(bottom)
	}
	if left != nil {
		leftObj = unwrap(left)
	}
	if right != nil {
		rightObj = unwrap(right)
	}
	if center != nil {
		centerObj = unwrap(center)
	}

	content := container.NewBorder(
		topObj,
		bottomObj,
		leftObj,
		rightObj,
		centerObj, // center fills remaining space
	)

	return wrap(content)
}
