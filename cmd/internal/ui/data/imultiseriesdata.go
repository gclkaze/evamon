package data

import (
	"time"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

type IMultiSeriesData interface {
	Vars() int
	Len() int
	MaxPoints() int

	// safe copies, aligned across all series
	ReadWindow(start, end int) (times []time.Time, values [][]int)

	// append one timestamp + N values (N == Vars())
	Append(at time.Time, nums []int)

	ApplyFilterChanges(changes []models.FilterComponentChange)
}
