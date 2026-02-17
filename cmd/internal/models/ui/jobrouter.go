package ui

import (
	"sync"
	"time"

	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

type JobRouter struct {
	mu   sync.RWMutex
	jobs map[string]*VariableContainer // jobID -> variables container
}

func NewJobRouter() *JobRouter {
	return &JobRouter{jobs: make(map[string]*VariableContainer)}
}

// Register registers a sink for (jobID, variable).
func (r *JobRouter) Register(jobID, variable string, w dport.EvaWidget) (remove func()) {
	r.mu.Lock()

	vc := r.jobs[jobID]
	if vc == nil {
		vc = NewVariableContainer()
		r.jobs[jobID] = vc
	}

	// Register to the variable container
	unsub := vc.Register(variable, w)

	r.mu.Unlock()

	// Unsubscribe closure removes from variable container and garbage-collects empty job containers.
	return func() {
		unsub()

		// Optional GC: if job container becomes empty, remove it.
		// This needs a method to check emptiness safely.
		r.mu.Lock()
		defer r.mu.Unlock()

		if vc2 := r.jobs[jobID]; vc2 != nil && vc2.IsEmpty() {
			delete(r.jobs, jobID)
		}
	}
}

// Push fans out to all widgets registered for (jobID, variable).
func (r *JobRouter) Push(jobID, variable string, t time.Time, v any) {
	r.mu.RLock()
	vc := r.jobs[jobID]
	r.mu.RUnlock()

	if vc == nil {
		return
	}
	vc.Push(variable, t, v)
}
