package ui

import (
	"sync"
	"time"

	"github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

type VariableContainer struct {
	mu      sync.RWMutex
	drawers map[string]map[port.DiagramWidget]struct{}
}

func NewVariableContainer() *VariableContainer {
	return &VariableContainer{drawers: make(map[string]map[port.DiagramWidget]struct{})}
}

func (inst *VariableContainer) Register(variableName string, drawer port.DiagramWidget /*port.Drawer*/) (remove func()) {
	inst.mu.Lock()

	set := inst.drawers[variableName]
	if set == nil {
		set = make(map[port.DiagramWidget]struct{})
		inst.drawers[variableName] = set
	}

	set[drawer] = struct{}{}

	inst.mu.Unlock()

	// ---- unsubscribe closure ----
	return func() {
		inst.mu.Lock()
		defer inst.mu.Unlock()

		if set := inst.drawers[variableName]; set != nil {
			delete(set, drawer)
			if len(set) == 0 {
				delete(inst.drawers, variableName)
			}
		}
	}
}

func (inst *VariableContainer) Push(varName string, t time.Time, v any) {
	inst.mu.RLock()

	set := inst.drawers[varName]

	// Copy to slice to avoid holding lock during callbacks
	drawers := make([]port.DiagramWidget, 0, len(set))
	for d := range set {
		drawers = append(drawers, d)
	}

	inst.mu.RUnlock()

	// Fan-out outside the lock
	for _, d := range drawers {
		d.Push(t, v)
	}
}
