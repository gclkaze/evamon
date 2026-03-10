package models

type MultiSetupItem struct {
	Variable     string    `json:"variable"` // key inside bundle (cpu/mem/gpu)
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	VariableType ValueType `json:"variableType"`

	DiagramStyle Style `json:"diagramStyle,omitempty"`
}
