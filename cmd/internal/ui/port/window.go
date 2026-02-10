package port

import "fyne.io/fyne/v2"

type ExecutionWindow interface {
	SetTitle(string)
	Show()
	Close()
	SetContent(fyne.CanvasObject)
	RunOnUI(func())

	// capabilities you need in WidgetService:
	SetResizable(resizable bool)  // true = user can resize
	Resize(width, height float32) // initial size
}
