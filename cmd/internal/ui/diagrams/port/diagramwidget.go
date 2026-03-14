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
}
