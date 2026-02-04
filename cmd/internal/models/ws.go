package models

// Generic envelope used by evamon <-> evacron
type WSMessage struct {
	Type  string      `json:"type"`
	ID    string      `json:"id,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}
