package models

import "time"

type EventType string

const (
	EventStarted EventType = "started"
	EventStdout  EventType = "stdout"
	EventStderr  EventType = "stderr"
	EventExited  EventType = "exited"
	EventError   EventType = "error"
)

type ScriptEvent struct {
	JobID string
	Type  EventType
	Time  time.Time

	Text     string
	ExitCode *int
}
