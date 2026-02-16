package port

import "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"

type ExecutionWindow interface {
	SetTitle(string)
	Show()
	Close()
	SetContent(port.Drawer) //(fyne.CanvasObject)
	RunOnUI(func())

	// capabilities you need in WidgetService:
	SetResizable(resizable bool)  // true = user can resize
	Resize(width, height float32) // initial size

	UpsertTab(tabID, title string)
	RemoveTab(tabID string)

	CommitTabs()
	AppendLog(tabID, line string)
	AssignTab(tabID string, draw port.Drawer)
	SetOnClosed(close func())
	//UpsertSeries(tabID, series string, points []Point)
}
