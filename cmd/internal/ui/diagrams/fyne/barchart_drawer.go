package fynediagrams

import (
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// BarChartDrawer draws the last N integer values as vertical bars.
type BarChartDrawer struct {
	maxPoints int
	width     float32
	height    float32

	root *fyne.Container

	values []int
	times  []time.Time

	barColor color.Color
}

func NewBarChartDrawer(maxPoints int, width, height float32) *BarChartDrawer {
	if maxPoints <= 0 {
		maxPoints = 50
	}
	if width <= 0 {
		width = 600
	}
	if height <= 0 {
		height = 240
	}

	d := &BarChartDrawer{
		maxPoints: maxPoints,
		width:     width,
		height:    height,
		barColor:  color.NRGBA{R: 80, G: 130, B: 255, A: 255},
	}
	d.root = container.NewWithoutLayout()
	d.root.Resize(fyne.NewSize(width, height))
	return d
}

func (d *BarChartDrawer) Root() fyne.CanvasObject { return d.root }

func (d *BarChartDrawer) Push(at time.Time, v int) {
	d.values = append(d.values, v)
	d.times = append(d.times, at)

	if len(d.values) > d.maxPoints {
		d.values = d.values[len(d.values)-d.maxPoints:]
		d.times = d.times[len(d.times)-d.maxPoints:]
	}

	d.redraw()
}

func (d *BarChartDrawer) redraw() {
	d.root.Objects = nil

	n := len(d.values)
	if n == 0 {
		d.root.Refresh()
		return
	}

	maxV := 0
	for _, v := range d.values {
		if v > maxV {
			maxV = v
		}
	}
	if maxV == 0 {
		maxV = 1
	}

	padding := float32(6)
	usableW := d.width - 2*padding
	slotW := usableW / float32(n)
	barW := float32(math.Max(float64(slotW-2), 2))

	for i, v := range d.values {
		h := (float32(v) / float32(maxV)) * (d.height - 2*padding)
		x := padding + float32(i)*slotW
		y := d.height - padding - h

		r := canvas.NewRectangle(d.barColor)
		r.Move(fyne.NewPos(x, y))
		r.Resize(fyne.NewSize(barW, h))
		d.root.Add(r)
	}

	d.root.Refresh()
}
