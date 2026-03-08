package fynerenderer

import "fyne.io/fyne/v2"

func NewChildWindow(title string, width, height float32) *FyneWindow {
	app := fyne.CurrentApp()
	w := app.NewWindow(title)
	w.Resize(fyne.NewSize(width, height))
	return NewFyneWindow(w)
}
