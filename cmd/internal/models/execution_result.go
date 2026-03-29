package models

type ExecutionResult struct {
	JobID   string `json:"jobId"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
