package port

import (
	"time"

	"fyne.io/fyne/v2"
)

type Drawer interface {
	Root() fyne.CanvasObject
}

type BoolDrawer interface {
	Drawer
	Push(at time.Time, v bool)
}

type IntDrawer interface {
	Drawer
	Push(at time.Time, v int)
}
