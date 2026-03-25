package port

import "github.com/gclkaze/evamon/cmd/internal/models"

type IDiagram interface {
	GetID() string
	SetID(ID string)
	GetName() string
	CollectVariables() []string
	GetDiagramOwner() models.DiagramOwner
	GetSetup() []models.SetupItem
	GetFilter() *models.Filter
	GetTriggerRules() []*models.TriggerRule
	GetType() models.DiagramType
}
