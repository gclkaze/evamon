package fynediagrams

import "fyne.io/fyne/v2"

func UI(fn func()) {
	fyne.Do(fn) // use DoAndWait if you must block
}
