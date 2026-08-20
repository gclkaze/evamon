package models

import "time"

type ExecutionLogEntry struct {
	Stream    string `json:"-"` // excluded from JSON
	Line      string `json:"-"` // excluded from JSON
	ParsedMsg string // extracted from {"msg":"..."} if JSON, else same as Line
	Timestamp time.Time
}

type SimpleExecutionLogEntry struct {
	ParsedMsg string // extracted from {"msg":"..."} if JSON, else same as Line
	Timestamp time.Time
}

func (s ExecutionLogEntry) ToSimpleExecutionLogEntry() *SimpleExecutionLogEntry {
	return &SimpleExecutionLogEntry{ParsedMsg: s.ParsedMsg, Timestamp: s.Timestamp}
}

type TriggerExecution struct {
	TriggerID  string
	FilePath   string
	StartedAt  time.Time
	FinishedAt *time.Time
	Success    *bool
	Lines      []ExecutionLogEntry
}

func NewTriggerExecution(triggerID, filePath string) *TriggerExecution {
	return &TriggerExecution{
		TriggerID: triggerID,
		FilePath:  filePath,
		StartedAt: time.Now().UTC(),
		Lines:     make([]ExecutionLogEntry, 0),
	}
}

func (e *TriggerExecution) AddLine(entry ExecutionLogEntry) {
	e.Lines = append(e.Lines, entry)
}

func (e *TriggerExecution) Finish(success bool) {
	now := time.Now().UTC()
	e.FinishedAt = &now
	e.Success = &success
}
