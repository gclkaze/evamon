package models

type ExecutionOutput struct {
	JobID    string `json:"jobId"`
	FilePath string `json:"filePath"`
	Stream   string `json:"stream"`
	Line     string `json:"line"`
}
