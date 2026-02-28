package data

import (
	"sync"
	"time"

	port "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type ringSeries struct {
	name string
	unit string

	vals []int
	ts   []time.Time

	head int // next write index
	size int // number of valid points (<= cap)
}

func (s *ringSeries) cap() int { return len(s.vals) }

func (s *ringSeries) append(v int, t time.Time) {
	if s.cap() == 0 {
		return
	}
	s.vals[s.head] = v
	s.ts[s.head] = t
	s.head = (s.head + 1) % s.cap()
	if s.size < s.cap() {
		s.size++
	}
}

func (s *ringSeries) len() int { return s.size }

func (s *ringSeries) latest() (int, time.Time, bool) {
	if s.size == 0 {
		return 0, time.Time{}, false
	}
	// last written index is head-1
	i := s.head - 1
	if i < 0 {
		i = s.cap() - 1
	}
	return s.vals[i], s.ts[i], true
}

// read window in logical order [0..size)
func (s *ringSeries) readWindow(start, end int) ([]int, []time.Time) {
	n := s.size
	if n == 0 {
		return nil, nil
	}
	if start < 0 {
		start = 0
	}
	if end <= 0 || end > n {
		end = n
	}
	if start >= end {
		return nil, nil
	}

	outV := make([]int, end-start)
	outT := make([]time.Time, end-start)

	// logical index 0 is the oldest element:
	// oldest physical index = head - size (wrapped)
	oldest := s.head - s.size
	for oldest < 0 {
		oldest += s.cap()
	}

	for k := start; k < end; k++ {
		pi := (oldest + k) % s.cap()
		outV[k-start] = s.vals[pi]
		outT[k-start] = s.ts[pi]
	}
	return outV, outT
}

type DataContainer struct {
	mu     sync.RWMutex
	series []*ringSeries
}

func NewDataContainer(maxPoints int, metas []port.SeriesMeta) *DataContainer {
	dc := &DataContainer{
		series: make([]*ringSeries, len(metas)),
	}
	for i, m := range metas {
		dc.series[i] = &ringSeries{
			name: m.Name,
			unit: m.Unit,
			vals: make([]int, maxPoints),
			ts:   make([]time.Time, maxPoints),
		}
	}
	return dc
}

func (dc *DataContainer) SeriesCount() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return len(dc.series)
}

func (dc *DataContainer) SeriesMeta(i int) port.SeriesMeta {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if i < 0 || i >= len(dc.series) {
		return port.SeriesMeta{}
	}
	s := dc.series[i]
	return port.SeriesMeta{Name: s.name, Unit: s.unit}
}

func (dc *DataContainer) Len(i int) int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if i < 0 || i >= len(dc.series) {
		return 0
	}
	return dc.series[i].len()
}

func (dc *DataContainer) Latest(i int) (int, time.Time, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if i < 0 || i >= len(dc.series) {
		return 0, time.Time{}, false
	}
	return dc.series[i].latest()
}

func (dc *DataContainer) ReadWindow(i int, start, end int) ([]int, []time.Time) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if i < 0 || i >= len(dc.series) {
		return nil, nil
	}
	return dc.series[i].readWindow(start, end)
}

func (dc *DataContainer) Append(i int, v int, t time.Time) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if i < 0 || i >= len(dc.series) {
		return
	}
	dc.series[i].append(v, t)
}
