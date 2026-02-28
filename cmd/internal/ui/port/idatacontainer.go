package port

import "time"

type SeriesMeta struct {
	Name string
	Unit string // optional
}

type IDataContainer interface {
	// Series list / mapping
	SeriesCount() int
	SeriesMeta(i int) SeriesMeta

	// Latest values
	Latest(i int) (v int, t time.Time, ok bool)

	// Range read (index window). Returns copies.
	// If end is 0 or end > Len(i), treat as Len(i).
	ReadWindow(i int, start, end int) (vals []int, ts []time.Time)

	// Length / bounds helpers
	Len(i int) int

	// Push data
	Append(i int, v int, t time.Time)
}
