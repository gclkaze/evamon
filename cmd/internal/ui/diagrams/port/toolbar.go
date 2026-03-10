package port

import (
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type DiagramToolbarFactory interface {
	Build(jobID string, d uport.IDiagram) uport.UIObject
}
