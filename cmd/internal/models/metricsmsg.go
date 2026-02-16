package models

type MetricsMsg struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Index string `json:"index"`
	Value any    `json:"value"`
}
