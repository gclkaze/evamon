package port

import (
	"time"

	"fyne.io/fyne/v2"
)

type DiagramWidget interface {
	fyne.CanvasObject
	Push(at time.Time, val any)
}
