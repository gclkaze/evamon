package port

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Drawer interface {
	Root() fyne.CanvasObject
	Push(at time.Time, v any)
	ToggleItem(*port.LegendItem)
	GetDataSeries() data.IMultiSeriesData

	Redraw()
}

type ResizableDrawer struct {
	widget.BaseWidget
	drawer Drawer
}

type BoolDrawer interface {
	ResizableDrawer
}

type IntDrawer interface {
	ResizableDrawer
}
