package services

import (
	"time"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

type TriggerExecutionBuilder struct {
	execution *models.TriggerExecution
}

func NewTriggerExecutionBuilder(triggerID, filePath string) *TriggerExecutionBuilder {
	return &TriggerExecutionBuilder{
		execution: models.NewTriggerExecution(triggerID, filePath),
	}
}

func (b *TriggerExecutionBuilder) AddLine(stream, line string) {
	b.execution.AddLine(models.ExecutionLogEntry{
		Stream:    stream,
		Line:      line,
		ParsedMsg: parseLogLine(line),
		Timestamp: time.Now().UTC(),
	})
}

func (b *TriggerExecutionBuilder) Finish(success bool) *models.TriggerExecution {
	b.execution.Finish(success)
	return b.execution
}
