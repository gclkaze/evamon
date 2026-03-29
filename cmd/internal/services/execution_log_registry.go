package services

import (
	"sync"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

const executionRingCapacity = 10

type ExecutionLogRegistry struct {
	mu   sync.RWMutex
	logs map[string]map[string]*models.RingBuffer[models.TriggerExecution]
}

func NewExecutionLogRegistry() *ExecutionLogRegistry {
	return &ExecutionLogRegistry{
		logs: make(map[string]map[string]*models.RingBuffer[models.TriggerExecution]),
	}
}

func (r *ExecutionLogRegistry) getOrCreate(diagramID, ruleID string) *models.RingBuffer[models.TriggerExecution] {
	if _, ok := r.logs[diagramID]; !ok {
		r.logs[diagramID] = make(map[string]*models.RingBuffer[models.TriggerExecution])
	}
	if _, ok := r.logs[diagramID][ruleID]; !ok {
		r.logs[diagramID][ruleID] = models.NewRingBuffer[models.TriggerExecution](executionRingCapacity)
	}
	return r.logs[diagramID][ruleID]
}

func (r *ExecutionLogRegistry) Push(diagramID, ruleID string, exec *models.TriggerExecution) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getOrCreate(diagramID, ruleID).Push(exec)
}

func (r *ExecutionLogRegistry) Get(diagramID, ruleID string) []*models.TriggerExecution {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.logs[diagramID]; ok {
		if ring, ok := d[ruleID]; ok {
			return ring.All()
		}
	}
	return nil
}

func (r *ExecutionLogRegistry) Latest(diagramID, ruleID string) *models.TriggerExecution {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.logs[diagramID]; ok {
		if ring, ok := d[ruleID]; ok {
			return ring.Latest()
		}
	}
	return nil
}

// Snapshot returns a point-in-time copy of all stored executions as
// diagramID → ruleID → []TriggerExecution slices. The returned data is
// independent of the registry's internal state.
func (r *ExecutionLogRegistry) Snapshot() map[string]map[string][]*models.TriggerExecution {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]map[string][]*models.TriggerExecution, len(r.logs))
	for diagID, rules := range r.logs {
		out[diagID] = make(map[string][]*models.TriggerExecution, len(rules))
		for ruleID, ring := range rules {
			out[diagID][ruleID] = ring.All()
		}
	}
	return out
}
