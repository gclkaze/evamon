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
	mu      sync.Mutex
	entries map[string]*triggerStreamEntry
}

func newTriggerStreamTracker() *triggerStreamTracker {
	return &triggerStreamTracker{entries: make(map[string]*triggerStreamEntry)}
}

func (t *triggerStreamTracker) register(msg *models.TriggerOperationMsg, port int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[msg.ID] = &triggerStreamEntry{
		Msg:    msg,
		Port:   port,
		Status: models.TriggerOperationStatusPending,
	}
}

func (t *triggerStreamTracker) updateStatus(msgID string, status models.TriggerOperationStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e, ok := t.entries[msgID]; ok {
		e.Status = status
	}
}

func (t *triggerStreamTracker) remove(msgID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, msgID)
}
