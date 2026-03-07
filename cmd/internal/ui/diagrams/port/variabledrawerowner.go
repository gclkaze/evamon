package port

import dport "github.com/gclkaze/evamon/cmd/internal/ui/port"

type VariableDrawerOwner interface {
	RegisterVariableDrawerUnsubscriber(variableName string, drawer DiagramWidget)
	GetRenderer() dport.Renderer
}
