package fynediagrams

import "fyne.io/fyne/v2"

func UI(fn func()) {
	fyne.Do(fn) // use DoAndWait if you must block
}

// zoomStep is the number of data points added or removed per zoom action.
// zoomMin is the smallest allowed visible window (points).
const (
	zoomStep = 2
	zoomMin  = 1
)
