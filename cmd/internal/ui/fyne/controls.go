package fynerenderer

import (
	fyne2 "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Controls struct{}

func (Controls) IconButton(icon uport.ToolbarIcon, onClick func()) uport.UIObject {
	var res fyne2.Resource

	switch icon {
	case uport.IconMaximize:
		res = theme.ViewFullScreenIcon()
	case uport.IconDownload:
		res = theme.DownloadIcon()
	case uport.IconFilters:
		res = theme.ComputerIcon()
	default:
		res = theme.QuestionIcon()
	}

	return wrap(widget.NewButtonWithIcon("", res, onClick))
}

func (Controls) Check(label string, checked bool, onChanged func(bool)) uport.UIObject {
	ch := widget.NewCheck(label, onChanged)
	ch.SetChecked(checked)
	return wrap(ch)
}

func (Controls) IconMenu(icon uport.ToolbarIcon, items []uport.MenuItem) uport.UIObject {
	var res fyne2.Resource

	switch icon {
	case uport.IconDownload:
		res = theme.DownloadIcon()
	case uport.IconFilters:
		res = theme.ComputerIcon()
	default:
		res = theme.MenuDropDownIcon()
	}

	btn := widget.NewButtonWithIcon("", res, nil)

	menuItems := make([]*fyne2.MenuItem, 0, len(items))
	for _, it := range items {
		item := it
		menuItems = append(menuItems, fyne2.NewMenuItem(item.Label, func() {
			if item.Action != nil {
				item.Action()
			}
		}))
	}

	menu := fyne2.NewMenu("", menuItems...)

	btn.OnTapped = func() {
		canvas := fyne2.CurrentApp().Driver().CanvasForObject(btn)
		if canvas == nil {
			return
		}
		widget.ShowPopUpMenuAtPosition(menu, canvas, btn.Position())
	}

	return wrap(btn)
}
