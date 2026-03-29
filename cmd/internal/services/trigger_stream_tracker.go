package services

import (
	"sync"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// triggerStreamEntry holds the trigger message and the stream socket assigned
// to it by Evacron, along with the current lifecycle status.
type triggerStreamEntry struct {
	Msg    *models.TriggerOperationMsg
	Port   int
	Status models.TriggerOperationStatus
}

// triggerStreamTracker maps TriggerOperationMsg.ID → entry so that every
// in-flight trigger can be inspected or cancelled independently.
type triggerStreamTracker struct {
	mu          sync.Mutex
	entries     map[string]*triggerStreamEntry
	coordinator *TriggerExecutionCoordinator
}

func newTriggerStreamTracker(coordinator *TriggerExecutionCoordinator) *triggerStreamTracker {
	return &triggerStreamTracker{
		entries:     make(map[string]*triggerStreamEntry),
		coordinator: coordinator,
	}
}

func (t *triggerStreamTracker) register(msg *models.TriggerOperationMsg, port int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[msg.ID] = &triggerStreamEntry{
		Msg:    msg,
		Port:   port,
		Status: models.TriggerOperationStatusPending,
	}
	t.coordinator.OnStart(msg)
}

func (t *triggerStreamTracker) GetMessage(msgID string) *models.TriggerOperationMsg {
	if e, ok := t.entries[msgID]; ok {
		return e.Msg
	}

	return nil
}

func (t *triggerStreamTracker) recordLine(triggerID string, output models.ExecutionOutput) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.coordinator.OnLine(triggerID, output)
}

func (t *triggerStreamTracker) recordError(triggerID, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	output := models.ExecutionOutput{
		JobID: triggerID,
		Line:  errMsg,
	}
	t.coordinator.OnLine(triggerID, output)
}

func (t *triggerStreamTracker) completeWithResult(msg *models.TriggerOperationMsg, success bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.coordinator.OnDone(msg, success)
	t.entries[msg.ID].Status = models.TriggerOperationStatusDone
}

func (t *triggerStreamTracker) remove(msgID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, msgID)
}

func (t *triggerStreamTracker) updateStatus(msgID string, status models.TriggerOperationStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e, ok := t.entries[msgID]; ok {
		e.Status = status
	}
}
