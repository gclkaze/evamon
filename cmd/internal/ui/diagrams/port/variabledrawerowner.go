package port

type VariableDrawerOwner interface {
	RegisterVariableDrawerUnsubscriber(variableName string, drawer DiagramWidget)
}
