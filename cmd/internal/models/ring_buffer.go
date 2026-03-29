package models

import "sync"

type RingBuffer[T any] struct {
	mu       sync.RWMutex
	items    []*T
	capacity int
	head     int
	count    int
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		items:    make([]*T, capacity),
		capacity: capacity,
	}
}

func (r *RingBuffer[T]) Push(item *T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[r.head] = item
	r.head = (r.head + 1) % r.capacity
	if r.count < r.capacity {
		r.count++
	}
}

func (r *RingBuffer[T]) All() []*T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ordered()
}

func (r *RingBuffer[T]) ordered() []*T {
	result := make([]*T, 0, r.count)
	start := (r.head - r.count + r.capacity) % r.capacity
	for i := 0; i < r.count; i++ {
		result = append(result, r.items[(start+i)%r.capacity])
	}
	return result
}

func (r *RingBuffer[T]) Latest() *T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.count == 0 {
		return nil
	}
	idx := (r.head - 1 + r.capacity) % r.capacity
	return r.items[idx]
}
