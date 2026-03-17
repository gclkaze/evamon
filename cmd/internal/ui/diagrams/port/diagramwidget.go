package port

import (
	"time"

	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type EvaWidget interface {
	Push(at time.Time, val any)
	//ApplyFilterChanges([]data.FilterComponentChange)
}

type DiagramWidget interface {
	uport.UIObject
	EvaWidget
	GetDataSeries() data.IMultiSeriesData

	// ZoomIn narrows the visible time window by one zoom step.
	// ZoomOut widens it. Both are no-ops when zooming is not applicable.
	ZoomIn()
	ZoomOut()

	// LastUpdatedLabel returns a UIObject that displays the timestamp of the
	// most recent Push call. The text is updated automatically on each Push.
	LastUpdatedLabel() uport.UIObject
}
