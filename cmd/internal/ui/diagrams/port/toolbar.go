package port

import (
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
)

type DiagramToolbarFactory interface {
	Build(jobID string, d *vp.Diagram) uport.UIObject
}
