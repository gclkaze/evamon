package port

type ToolbarIcon string

const (
	IconMaximize ToolbarIcon = "maximize"
	IconDownload ToolbarIcon = "download"
	IconFilters  ToolbarIcon = "filters"
)

type MenuItem struct {
	Label  string
	Action func()
}

type Controls interface {
	IconButton(icon ToolbarIcon, onClick func()) UIObject
	IconMenu(icon ToolbarIcon, items []MenuItem) UIObject
}
