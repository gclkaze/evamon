package models

import (
	"encoding/json"
	"fmt"
)

// Generic envelope used by evamon <-> evacron
type WSMessage struct {
	Type  string          `json:"type"`
	ID    string          `json:"id,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

func (m *WSMessage) DataAsBool() (bool, error) {
	if m.Data == nil {
		return false, nil
	}

	var b bool
	if err := json.Unmarshal(m.Data, &b); err != nil {
		return false, fmt.Errorf("data is not a boolean: %w", err)
	}

	return b, nil
}

func (m *WSMessage) DataAsBoolMap() (map[string]bool, error) {
	if m.Data == nil {
		return nil, nil
	}

	b := make(map[string]bool)
	if err := json.Unmarshal(m.Data, &b); err != nil {
		return nil, fmt.Errorf("data is not a map of booleans: %w", err)
	}

	return b, nil
}
