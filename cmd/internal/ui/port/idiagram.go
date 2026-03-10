package port

import "github.com/gclkaze/evamon/cmd/internal/models"

type IDiagram interface {
	GetName() string
	CollectVariables() []string
	GetDiagramOwner() models.DiagramOwner
	GetSetup() []models.SetupItem
	GetType() models.DiagramType
}
