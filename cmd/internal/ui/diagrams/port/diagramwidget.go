package port

import (
	"time"

	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type EvaWidget interface {
	Push(at time.Time, val any)
}

type DiagramWidget interface {
	uport.UIObject
	EvaWidget
}
