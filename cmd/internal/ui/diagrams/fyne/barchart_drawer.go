package fynediagrams

import (
	"fmt"
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
	bgColor  color.Color

	bg *canvas.Rectangle
}

func (d *BarChartDrawer) Redraw() {
	d.redrawWithAxis()
}
func NewBarChartDrawer(maxPoints int, width, height float32, barColor color.Color, bgColor color.Color) *BarChartDrawer {
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
		barColor:  barColor,
		bgColor:   bgColor,
	}
	d.root = container.NewWithoutLayout()
	d.root.Resize(fyne.NewSize(width, height))

	d.bg = canvas.NewRectangle(d.bgColor)
	//d.bg.Resize(fyne.NewSize(d.width, d.height))
	d.bg.Move(fyne.NewPos(0, 0))
	d.root.Add(d.bg)
	return d
}

func (d *BarChartDrawer) Root() fyne.CanvasObject { return d.root }

func (d *BarChartDrawer) Push(at time.Time, val any) {
	UI(func() {
		v := 0
		switch x := val.(type) {
		case int:
			v = x
		case float64:
			v = int(x)
		case float32:
			v = int(x)
		default:
			return
		}

		if v < 0 {
			v = 0
		}

		d.values = append(d.values, v)
		d.times = append(d.times, at)

		if len(d.values) > d.maxPoints {
			d.values = d.values[len(d.values)-d.maxPoints:]
			d.times = d.times[len(d.times)-d.maxPoints:]
		}

		d.redrawWithAxis()
	})
}

/*
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
}*/

func (d *BarChartDrawer) redraw() {
	// Use the *actual* size allocated by the window/layout.
	sz := d.root.Size()
	w, h := sz.Width, sz.Height
	if w <= 1 || h <= 1 {
		// Not laid out yet; nothing meaningful to draw.
		return
	}

	// Resize background to fill allocated area.
	d.bg.Resize(fyne.NewSize(w, h))

	// Clear bars but keep background.
	d.root.Objects = d.root.Objects[:0]
	d.root.Add(d.bg)

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
	if maxV <= 0 {
		maxV = 1
	}

	padding := float32(8)
	usableW := w - 2*padding
	usableH := h - 2*padding
	if usableW <= 1 || usableH <= 1 {
		d.root.Refresh()
		return
	}

	slotW := usableW / float32(n)
	barW := float32(math.Max(float64(slotW-2), 2))

	for i, v := range d.values {
		bh := (float32(v) / float32(maxV)) * usableH
		if bh < 1 {
			bh = 1 // keep visible
		}

		x := padding + float32(i)*slotW
		y := h - padding - bh

		r := canvas.NewRectangle(d.barColor)
		r.Move(fyne.NewPos(x, y))
		r.Resize(fyne.NewSize(barW, bh))
		d.root.Add(r)
	}

	d.root.Refresh()
}

func (d *BarChartDrawer) redrawWithAxis() {
	sz := d.root.Size()
	w, h := sz.Width, sz.Height
	if w <= 1 || h <= 1 {
		return
	}

	d.bg.Resize(fyne.NewSize(w, h))

	// Clear everything but keep background.
	d.root.Objects = d.root.Objects[:0]
	d.root.Add(d.bg)

	n := len(d.values)
	if n == 0 {
		d.root.Refresh()
		return
	}

	// ---- styling ----
	axisColor := color.NRGBA{R: 170, G: 170, B: 170, A: 255}
	gridColor := color.NRGBA{R: 255, G: 255, B: 255, A: 28} // subtle
	labelColor := color.NRGBA{R: 225, G: 225, B: 225, A: 255}

	// IMPORTANT: leave enough room for y labels (k/M/B) and tick marks.
	marginLeft := float32(78)
	marginRight := float32(12)
	marginTop := float32(12)
	marginBottom := float32(34)

	plotX0 := marginLeft
	plotY0 := marginTop
	plotX1 := w - marginRight
	plotY1 := h - marginBottom

	plotW := plotX1 - plotX0
	plotH := plotY1 - plotY0
	if plotW <= 2 || plotH <= 2 {
		d.root.Refresh()
		return
	}

	// ---- data range (with headroom) ----
	maxV := 0
	for _, v := range d.values {
		if v > maxV {
			maxV = v
		}
	}
	if maxV <= 0 {
		maxV = 1
	}

	// add ~10% headroom so top bar doesn’t touch ceiling
	maxWithHeadroom := float64(maxV) * 1.10
	ticks := computeNiceTicks(0, maxWithHeadroom, 6) // target ~6 ticks

	yMin := ticks.Min
	yMax := ticks.Max
	if yMax <= yMin {
		yMax = yMin + 1
	}

	// ---- axes ----
	yAxis := canvas.NewLine(axisColor)
	yAxis.Position1 = fyne.NewPos(plotX0, plotY0)
	yAxis.Position2 = fyne.NewPos(plotX0, plotY1)
	d.root.Add(yAxis)

	xAxis := canvas.NewLine(axisColor)
	xAxis.Position1 = fyne.NewPos(plotX0, plotY1)
	xAxis.Position2 = fyne.NewPos(plotX1, plotY1)
	d.root.Add(xAxis)

	// ---- horizontal grid + y labels ----
	tickLen := float32(6)

	for _, tv := range ticks.Ticks {
		frac := float32((tv - yMin) / (yMax - yMin)) // 0..1
		frac = clamp01(frac)
		y := plotY1 - frac*plotH

		// grid line across plot
		grid := canvas.NewLine(gridColor)
		grid.Position1 = fyne.NewPos(plotX0, y)
		grid.Position2 = fyne.NewPos(plotX1, y)
		d.root.Add(grid)

		// tick mark
		tick := canvas.NewLine(axisColor)
		tick.Position1 = fyne.NewPos(plotX0-tickLen, y)
		tick.Position2 = fyne.NewPos(plotX0, y)
		d.root.Add(tick)

		// label
		lbl := canvas.NewText(formatCompact(tv), labelColor)
		lbl.TextSize = 11
		lbl.Alignment = fyne.TextAlignTrailing
		lbl.Refresh()
		ls := lbl.MinSize()
		lbl.Move(fyne.NewPos(plotX0-tickLen-6-ls.Width, y-ls.Height/2))
		d.root.Add(lbl)
	}

	// ---- bars ----
	slotW := plotW / float32(n)

	// Keep bars readable: max 80% slot width, min 2px.
	barW := float32(math.Max(float64(slotW*0.80), 2))

	for i, v := range d.values {
		// map v -> y using nice yMax (not raw maxV)
		val := float64(v)
		frac := float32((val - yMin) / (yMax - yMin))
		frac = clamp01(frac)

		bh := frac * plotH
		if bh < 1 {
			bh = 1
		}

		x := plotX0 + float32(i)*slotW + (slotW-barW)/2
		y := plotY1 - bh

		r := canvas.NewRectangle(d.barColor)
		r.Move(fyne.NewPos(x, y))
		r.Resize(fyne.NewSize(barW, bh))
		d.root.Add(r)
	}

	// ---- X labels (times) with overlap-aware density ----
	// ---- X labels (times) with overlap-aware density ----
	if len(d.times) == n && n > 0 {
		// Special-case: only one point -> one label, no division by (maxLabels-1)
		if n == 1 {
			t := d.times[0]
			lbl := canvas.NewText(t.Format("15:04:05"), labelColor)
			lbl.TextSize = 11
			lbl.Alignment = fyne.TextAlignCenter
			lbl.Refresh()

			cx := plotX0 + plotW/2
			ls := lbl.MinSize()
			lbl.Move(fyne.NewPos(cx-ls.Width/2, plotY1+6))
			d.root.Add(lbl)

			// optional: small vertical tick on x-axis
			xTick := canvas.NewLine(axisColor)
			xTick.Position1 = fyne.NewPos(cx, plotY1)
			xTick.Position2 = fyne.NewPos(cx, plotY1+4)
			d.root.Add(xTick)
		} else {
			// Estimate how many labels fit: use a sample label width.
			sample := canvas.NewText("88:88:88", labelColor)
			sample.TextSize = 11
			sample.Refresh()
			sampleW := sample.MinSize().Width

			// Aim for at least sampleW+8 spacing between labels.
			minSpacing := sampleW + 8
			maxLabels := int(plotW / minSpacing)

			// clamp
			if maxLabels < 2 {
				maxLabels = 2
			}
			if maxLabels > 6 {
				maxLabels = 6
			}
			if maxLabels > n {
				maxLabels = n
			}
			// here: n>=2 => maxLabels>=2 => denom >= 1
			den := float64(maxLabels - 1)

			for k := 0; k < maxLabels; k++ {
				idx := int(math.Round(float64(k) * float64(n-1) / den))
				if idx < 0 {
					idx = 0
				}
				if idx > n-1 {
					idx = n - 1
				}

				t := d.times[idx]

				lbl := canvas.NewText(t.Format("15:04:05"), labelColor)
				lbl.TextSize = 11
				lbl.Alignment = fyne.TextAlignCenter
				lbl.Refresh()

				cx := plotX0 + (float32(idx)+0.5)*slotW
				ls := lbl.MinSize()
				lbl.Move(fyne.NewPos(cx-ls.Width/2, plotY1+6))
				d.root.Add(lbl)

				// optional: small vertical tick on x-axis
				xTick := canvas.NewLine(axisColor)
				xTick.Position1 = fyne.NewPos(cx, plotY1)
				xTick.Position2 = fyne.NewPos(cx, plotY1+4)
				d.root.Add(xTick)
			}
		}
	}
	d.root.Refresh()
}

func formatInt(v int) string {
	// keep this simple; customize if you want (e.g., 1.2k)
	return fmt.Sprintf("%d", v)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func niceNum(x float64, round bool) float64 {
	// “Nice numbers” for ticks: 1, 2, 5, 10 * 10^n
	if x <= 0 {
		return 1
	}
	exp := math.Floor(math.Log10(x))
	f := x / math.Pow(10, exp)

	var nf float64
	if round {
		switch {
		case f < 1.5:
			nf = 1
		case f < 3:
			nf = 2
		case f < 7:
			nf = 5
		default:
			nf = 10
		}
	} else {
		switch {
		case f <= 1:
			nf = 1
		case f <= 2:
			nf = 2
		case f <= 5:
			nf = 5
		default:
			nf = 10
		}
	}
	return nf * math.Pow(10, exp)
}

type TickSpec struct {
	Min, Max float64
	Step     float64
	Ticks    []float64
}

func computeNiceTicks(minV, maxV float64, targetTicks int) TickSpec {
	if targetTicks < 2 {
		targetTicks = 2
	}
	if maxV <= minV {
		maxV = minV + 1
	}

	rng := niceNum(maxV-minV, false)
	step := niceNum(rng/float64(targetTicks-1), true)

	niceMin := math.Floor(minV/step) * step
	niceMax := math.Ceil(maxV/step) * step

	// avoid infinite loop if step is tiny
	if step <= 0 {
		step = 1
	}

	var ticks []float64
	for v := niceMin; v <= niceMax+0.5*step; v += step {
		ticks = append(ticks, v)
	}
	return TickSpec{Min: niceMin, Max: niceMax, Step: step, Ticks: ticks}
}

func formatCompact(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", v/1_000_000_000)
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", v/1_000)
	default:
		// for integer-y values, don’t show decimals
		if v == math.Trunc(v) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.2f", v)
	}
}

func clamp01(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
