package port

type ExecutionWindow interface {
	SetTitle(string)
	Show()
	Close()
	SetContent(content UIObject) //(fyne.CanvasObject)
	RunOnUI(func())

	// capabilities you need in WidgetService:
	SetResizable(resizable bool)  // true = user can resize
	Resize(width, height float32) // initial size

	UpsertTab(tabID, title string)
	RemoveTab(tabID string)

	CommitTabs()
	AppendLog(tabID, line string)
	AssignTab(tabID string, content UIObject)
	SetOnClosed(close func())
	//UpsertSeries(tabID, series string, points []Point)
}
