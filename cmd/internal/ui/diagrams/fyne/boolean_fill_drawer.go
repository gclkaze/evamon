package fynediagrams

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// BoolFillDrawer fills its area with a color based on the last boolean value.
type BoolFillDrawer struct {
	rect       *canvas.Rectangle
	root       fyne.CanvasObject
	trueColor  color.Color
	falseColor color.Color
}

func NewBoolFillDrawer(trueColor, falseColor color.Color) *BoolFillDrawer {
	r := canvas.NewRectangle(falseColor)

	// NewMax makes it expand to fill available space and resize nicely.
	root := container.NewMax(r)

	return &BoolFillDrawer{
		rect:       r,
		root:       root,
		trueColor:  trueColor,
		falseColor: falseColor,
	}
}

func (d *BoolFillDrawer) Root() fyne.CanvasObject { return d.root }

// Push updates ONLY the last value (fills the panel green/red etc).
func (d *BoolFillDrawer) Push(_ time.Time, v bool) {
	if v {
		d.rect.FillColor = d.trueColor
	} else {
		d.rect.FillColor = d.falseColor
	}
	d.rect.Refresh()
}
