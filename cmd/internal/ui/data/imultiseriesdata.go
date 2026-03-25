package data

import (
	"time"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// TriggerSendFunc is called when a trigger rule fires (rising-edge transition).
// ruleID is the rule's OriginalComponentID; files is the ordered list of .eva paths.
type TriggerSendFunc func(ruleID string, files []string)

type IMultiSeriesData interface {
	Vars() int
	Len() int
	MaxPoints() int

	// safe copies, aligned across all series
	ReadWindow(start, end int) (times []time.Time, values [][]int)

	// append one timestamp + N values (N == Vars())
	Append(at time.Time, nums []int)

	ApplyFilterChanges(changes []models.FilterComponentChange)

	// SetTriggerSender injects the function that will be called (in a goroutine)
	// when a trigger rule fires. Passing nil disables trigger evaluation.
	SetTriggerSender(fn TriggerSendFunc)
}
