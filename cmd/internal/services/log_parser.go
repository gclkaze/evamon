package services

import (
	"encoding/json"
)

type evaLogLine struct {
	Msg string `json:"msg"`
}

func parseLogLine(raw string) string {
	var parsed evaLogLine
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return raw
	}
	if parsed.Msg == "" {
		return raw
	}
	return parsed.Msg
}
