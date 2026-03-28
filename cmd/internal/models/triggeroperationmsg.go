package models

import (
	"encoding/json"

	"github.com/gclkaze/evamon/pkg/utils"
)

type TriggerOperationStatus string

const (
	TriggerOperationStatusPending TriggerOperationStatus = "PENDING"
	TriggerOperationStatusAck     TriggerOperationStatus = "ACK"
	TriggerOperationStatusDone    TriggerOperationStatus = "DONE"
)

// Existing TriggerOperationMsg struct
type TriggerOperationMsg struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	JobID     string                 `json:"jobId"`
	RuleID    string                 `json:"ruleId"`
	Files     []string               `json:"files"`
	Status    TriggerOperationStatus `json:"status"`
	DiagramID string                 `json:"diagramID"`
}

// New method to convert TriggerOperationMsg to WSMessage
func (m *TriggerOperationMsg) ToWSMessage() (*WSMessage, error) {
	data, err := json.Marshal(m) // Marshal TriggerOperationMsg to JSON
	if err != nil {
		return nil, err // Return error if marshaling fails
	}

	return &WSMessage{
		Type:  m.Type, // Set the type as needed
		ID:    m.ID,   // Set the ID from TriggerOperationMsg
		Data:  data,   // Set the marshaled data
		Error: "",     // Set error to empty if no error
	}, nil
}

// Example usage of NewTriggerOperationMsg
func NewTriggerOperationMsg(jobID, ruleID, diagramID string, files []string) *TriggerOperationMsg {
	return &TriggerOperationMsg{
		ID:        utils.GetRandomString(),
		Type:      "job.trigger",
		JobID:     jobID,
		RuleID:    ruleID,
		Files:     files,
		DiagramID: diagramID,
		Status:    TriggerOperationStatusPending,
	}
}
