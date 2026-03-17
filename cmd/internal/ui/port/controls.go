package port

type ToolbarIcon string

const (
	IconMaximize ToolbarIcon = "maximize"
	IconDownload ToolbarIcon = "download"
	IconFilters  ToolbarIcon = "filters"
	IconZoomIn   ToolbarIcon = "zoom_in"
	IconZoomOut  ToolbarIcon = "zoom_out"
)

type MenuItem struct {
	Label  string
	Action func()
}

type Controls interface {
	IconButton(icon ToolbarIcon, onClick func()) UIObject
	IconMenu(icon ToolbarIcon, items []MenuItem) UIObject
	Check(label string, checked bool, onChanged func(bool)) UIObject
}
