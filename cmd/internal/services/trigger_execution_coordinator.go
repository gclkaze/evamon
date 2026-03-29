package services

import (
	"github.com/gclkaze/evamon/cmd/internal/models"
)

type TriggerExecutionCoordinator struct {
	registry *ExecutionLogRegistry
	active   map[string]*TriggerExecutionBuilder // map[TriggerID]→builder
}

func NewTriggerExecutionCoordinator(registry *ExecutionLogRegistry) *TriggerExecutionCoordinator {
	return &TriggerExecutionCoordinator{
		registry: registry,
		active:   make(map[string]*TriggerExecutionBuilder),
	}
}

func (c *TriggerExecutionCoordinator) OnStart(msg *models.TriggerOperationMsg) {
	builder := NewTriggerExecutionBuilder(msg.ID, firstFile(msg.Files))
	c.active[msg.ID] = builder
}

func (c *TriggerExecutionCoordinator) OnLine(triggerID string, output models.ExecutionOutput) {
	if b, ok := c.active[triggerID]; ok {
		b.AddLine(output.Stream, output.Line)
	}
}

func (c *TriggerExecutionCoordinator) OnDone(msg *models.TriggerOperationMsg) {
	b, ok := c.active[msg.ID]
	if !ok {
		return
	}
	success := msg.Status == models.TriggerOperationStatusDone
	exec := b.Finish(success)
	c.registry.Push(msg.DiagramID, msg.RuleID, exec)
	delete(c.active, msg.ID)
}

func firstFile(files []string) string {
	if len(files) > 0 {
		return files[0]
	}
	return ""
}
