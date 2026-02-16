package port

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Drawer interface {
	Root() fyne.CanvasObject
	Push(at time.Time, v any)
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
