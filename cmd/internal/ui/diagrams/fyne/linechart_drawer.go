package fynediagrams

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	port "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// LineChartDrawer draws multi-series points into a Raster and overlays axis labels.
// Multi-series behavior matches your BarChartDrawer:
//   - vars == 1: Push accepts int/float values
//   - vars > 1: Push expects a JSON array string (e.g. "[1,2,3]") and unmarshals to []int
//
// X-axis labels:
//   - Always shows first and last timestamps (properly aligned).
//   - Adds intermediate tick labels (2..maxXTicks, overlap-aware) with reused objects.
//   - Uses cached sample label width (no per-refresh allocations for measurement).
type LineChartDrawer struct {
	mu sync.Mutex

	opts port.LineChartOptions

	root *fyne.Container

	// Raster plot (anti-aliased drawing happens here)
	raster *canvas.Raster

	// Data
	values [][]int
	times  []time.Time
	vars   int

	// Cached scale
	minV int
	maxV int

	// Overlay labels
	yMinT *canvas.Text
	yMaxT *canvas.Text
	xMinT *canvas.Text
	xMaxT *canvas.Text

	// Intermediate X tick labels + tick lines (reused)
	xTicks     []*canvas.Text
	xTickLines []*canvas.Line
	maxXTicks  int // inclusive count (including endpoints) used for density calculation; we render maxXTicks-2 mid labels max.

	// Cached label measurement for overlap-aware density
	xLabelTextSize float32
	xLabelSampleW  float32
	xLabelCacheOK  bool

	// Last known size
	size fyne.Size

	img *image.RGBA

	shownIndices map[int]bool
}

func NewLineChartDrawer(opts port.LineChartOptions, initialWidth, initialHeight float32) *LineChartDrawer {
	if opts.MaxPoints <= 1 {
		opts.MaxPoints = 50
	}
	if opts.YPadRatio < 0 {
		opts.YPadRatio = 0
	}
	if opts.GridX < 0 {
		opts.GridX = 0
	}
	if opts.GridY < 0 {
		opts.GridY = 0
	}
	if initialWidth <= 0 {
		initialWidth = 800
	}
	if initialHeight <= 0 {
		initialHeight = 260
	}

	d := &LineChartDrawer{
		opts:           opts,
		size:           fyne.NewSize(initialWidth, initialHeight),
		xLabelTextSize: 10, // keep consistent with your existing label sizes
		maxXTicks:      6,  // density cap (similar to bar chart)
	}

	d.vars = len(opts.Variables)
	if d.vars <= 0 {
		d.vars = 1
	}

	d.shownIndices = map[int]bool{}
	for i := range d.vars {
		d.shownIndices[i] = true
	}
	d.values = make([][]int, d.vars)

	// Single raster strategy: generator returns d.img.
	d.raster = canvas.NewRaster(func(w, h int) image.Image {
		d.mu.Lock()
		defer d.mu.Unlock()

		if w <= 0 || h <= 0 {
			return image.NewRGBA(image.Rect(0, 0, 1, 1))
		}

		// Ensure backing buffer exists and matches requested size.
		if d.img == nil || d.img.Bounds().Dx() != w || d.img.Bounds().Dy() != h {
			d.img = image.NewRGBA(image.Rect(0, 0, w, h))
			fillRGBA(d.img, d.opts.Background)
		}

		return d.img
	})

	// Labels (overlay)
	d.yMinT = canvas.NewText("", opts.Axis)
	d.yMaxT = canvas.NewText("", opts.Axis)
	d.xMinT = canvas.NewText("", opts.Axis)
	d.xMaxT = canvas.NewText("", opts.Axis)

	d.yMinT.TextSize = d.xLabelTextSize
	d.yMaxT.TextSize = d.xLabelTextSize
	d.xMinT.TextSize = d.xLabelTextSize
	d.xMaxT.TextSize = d.xLabelTextSize

	d.yMinT.Alignment = fyne.TextAlignLeading
	d.yMaxT.Alignment = fyne.TextAlignLeading
	d.xMinT.Alignment = fyne.TextAlignLeading
	d.xMaxT.Alignment = fyne.TextAlignLeading

	// Precreate intermediate tick labels + lines (reused)
	for i := 0; i < d.maxXTicks; i++ {
		t := canvas.NewText("", opts.Axis)
		t.TextSize = d.xLabelTextSize
		t.Alignment = fyne.TextAlignCenter
		t.Hide()
		d.xTicks = append(d.xTicks, t)

		ln := canvas.NewLine(opts.Axis)
		ln.StrokeWidth = 1
		ln.Hide()
		d.xTickLines = append(d.xTickLines, ln)
	}

	// Root with absolute positioning (include tick objects)
	objects := []fyne.CanvasObject{d.raster, d.yMinT, d.yMaxT, d.xMinT, d.xMaxT}
	for i := 0; i < d.maxXTicks; i++ {
		objects = append(objects, d.xTickLines[i], d.xTicks[i])
	}
	d.root = container.NewWithoutLayout(objects...)
	d.root.Resize(d.size)

	d.raster.Move(fyne.NewPos(0, 0))
	d.raster.Resize(d.size)

	// Initial draw
	d.refreshLocked()

	return d
}

func (d *LineChartDrawer) variableShown(i int) bool {
	return d.shownIndices[i]
}

func (d *LineChartDrawer) ToggleItem(it *uport.LegendItem) {
	if len(d.shownIndices) == 1 {
		return
	}
	d.shownIndices[it.Index] = !d.shownIndices[it.Index]
	d.refreshLocked()

}
func (d *LineChartDrawer) Object() fyne.CanvasObject { return d.root }

// InvalidateLabelCache should be called when font metrics affecting label width change,
// e.g. if you change d.xLabelTextSize or app theme/font.
func (d *LineChartDrawer) InvalidateLabelCache() {
	d.xLabelCacheOK = false
	d.xLabelSampleW = 0
}

func (d *LineChartDrawer) Resize(width, height float32) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if width <= 0 || height <= 0 {
		return
	}
	d.size = fyne.NewSize(width, height)
	d.root.Resize(d.size)

	d.raster.Move(fyne.NewPos(0, 0))
	d.raster.Resize(d.size)

	d.refreshLocked()
}

func (d *LineChartDrawer) SetOptions(opts port.LineChartOptions) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if opts.MaxPoints <= 1 {
		opts.MaxPoints = d.opts.MaxPoints
	}
	if opts.YPadRatio < 0 {
		opts.YPadRatio = d.opts.YPadRatio
	}
	if opts.GridX < 0 {
		opts.GridX = 0
	}
	if opts.GridY < 0 {
		opts.GridY = 0
	}

	// Handle vars change
	newVars := len(opts.Variables)
	if newVars <= 0 {
		newVars = 1
	}
	if newVars != d.vars {
		// simplest/safest: reset data when series structure changes
		d.vars = newVars
		d.values = make([][]int, d.vars)
		d.times = nil
	}

	// If axis color changes, update label colors now (text objects are reused)
	d.opts = opts

	// Label measurement could change under theme/font; safest is to invalidate
	// if you mutate text size elsewhere; axis color doesn't change width though.
	// d.InvalidateLabelCache()

	// Trim if needed
	d.trimLocked()

	d.refreshLocked()
}

// Push adds a point and redraws.
// - If vars == 1: accepts int/float
// - If vars > 1: expects a JSON array string "[1,2,3]"
func (d *LineChartDrawer) Push(at time.Time, val any) {
	if d.vars == 1 {
		d.handleMonoVariableInput(at, val)
		return
	}

	s, ok := val.(string)
	if !ok {
		return
	}
	var nums []int
	if err := json.Unmarshal([]byte(s), &nums); err != nil {
		return
	}
	d.handleMultiVariableInput(at, &nums)
}

func (d *LineChartDrawer) handleMonoVariableInput(at time.Time, val any) {
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

		d.mu.Lock()
		defer d.mu.Unlock()

		d.values[0] = append(d.values[0], v)
		d.times = append(d.times, at)

		d.trimLocked()
		d.refreshLocked()
	})
}

func (d *LineChartDrawer) handleMultiVariableInput(at time.Time, nums *[]int) {
	UI(func() {
		if len(*nums) != d.vars {
			return
		}

		d.mu.Lock()
		defer d.mu.Unlock()

		// append timestamp ONCE per group
		d.times = append(d.times, at)

		// append one value per series
		for j := 0; j < d.vars; j++ {
			v := (*nums)[j]
			if v < 0 {
				v = 0
			}
			d.values[j] = append(d.values[j], v)
		}

		d.trimLocked()
		d.refreshLocked()
	})
}

func (d *LineChartDrawer) trimLocked() {
	points := len(d.times)
	if points <= d.opts.MaxPoints {
		return
	}

	start := points - d.opts.MaxPoints
	d.times = d.times[start:]

	for j := range d.values {
		if !d.variableShown(j) {
			continue
		}
		// normal case: each series has same length as times
		if len(d.values[j]) >= points {
			d.values[j] = d.values[j][start:]
			continue
		}
		// fallback safety
		if len(d.values[j]) > d.opts.MaxPoints {
			over := len(d.values[j]) - d.opts.MaxPoints
			d.values[j] = d.values[j][over:]
		}
	}
}

func (d *LineChartDrawer) pointCountSafeLocked() int {
	pc := len(d.times)
	for j := 0; j < d.vars; j++ {
		if !d.variableShown(j) {
			continue
		}

		if j >= len(d.values) {
			return 0
		}
		if len(d.values[j]) < pc {
			pc = len(d.values[j])
		}
	}
	if pc < 0 {
		return 0
	}
	return pc
}

func (d *LineChartDrawer) seriesColor(j int) color.Color {
	if j >= 0 && j < len(d.opts.Variables) && d.opts.Variables[j].VarColor != nil {
		return d.opts.Variables[j].VarColor
	}
	// fallback
	return d.opts.Axis
}

func (d *LineChartDrawer) sampleTimeLabelWidthLocked() float32 {
	if d.xLabelCacheOK && d.xLabelSampleW > 0 {
		return d.xLabelSampleW
	}

	sample := canvas.NewText("88:88:88", d.opts.Axis)
	sample.TextSize = d.xLabelTextSize
	sample.Alignment = fyne.TextAlignCenter
	sample.Refresh()

	d.xLabelSampleW = sample.MinSize().Width
	d.xLabelCacheOK = true

	if d.xLabelSampleW <= 0 {
		d.xLabelSampleW = 50
	}
	return d.xLabelSampleW
}

func (d *LineChartDrawer) refreshLocked() {
	w := int(d.size.Width)
	h := int(d.size.Height)
	if w <= 1 || h <= 1 {
		return
	}

	// Ensure backing buffer exists and matches size
	if d.img == nil || d.img.Bounds().Dx() != w || d.img.Bounds().Dy() != h {
		d.img = image.NewRGBA(image.Rect(0, 0, w, h))
	}

	// Clear background ONCE at the start
	fillRGBA(d.img, d.opts.Background)

	plotX0 := d.opts.PadL
	plotY0 := d.opts.PadT
	plotW := float32(w) - d.opts.PadL - d.opts.PadR
	plotH := float32(h) - d.opts.PadT - d.opts.PadB

	if plotW < 10 {
		plotW = 10
	}
	if plotH < 10 {
		plotH = 10
	}

	// Draw grid/axes
	if d.opts.ShowGrid {
		drawGrid(d.img, plotX0, plotY0, plotW, plotH, d.opts.Grid, d.opts.GridStroke, d.opts.GridX, d.opts.GridY)
	}
	if d.opts.ShowAxes {
		drawAxes(d.img, plotX0, plotY0, plotW, plotH, d.opts.Axis, d.opts.AxisStroke)
	}

	// -----------------------
	// Scale data across ALL series
	// -----------------------
	n := d.pointCountSafeLocked()

	minV, maxV := 0, 0
	if n > 0 {
		firstSet := false
		for j := 0; j < d.vars && !firstSet; j++ {
			if !d.variableShown(j) {
				continue
			}

			if j < len(d.values) && len(d.values[j]) >= n && n > 0 {
				minV, maxV = d.values[j][0], d.values[j][0]
				firstSet = true
			}
		}

		if firstSet {
			for j := 0; j < d.vars; j++ {
				if !d.variableShown(j) {
					continue
				}

				if j >= len(d.values) {
					break
				}
				series := d.values[j]
				if len(series) < n {
					continue
				}
				for i := 0; i < n; i++ {
					v := series[i]
					if v < minV {
						minV = v
					}
					if v > maxV {
						maxV = v
					}
				}
			}

			if minV == maxV {
				minV--
				maxV++
			}

			span := float64(maxV - minV)
			pad := int(math.Ceil(span * d.opts.YPadRatio))
			if pad < 1 {
				pad = 1
			}
			minV -= pad
			maxV += pad
		}
	}

	d.minV, d.maxV = minV, maxV

	// -----------------------
	// Plot mapping (shared X across series)
	// -----------------------
	toX := func(i int) float32 {
		if n <= 1 {
			return plotX0 + plotW/2
		}
		return plotX0 + (float32(i)/float32(n-1))*plotW
	}
	toY := func(v int) float32 {
		if maxV == minV {
			return plotY0 + plotH/2
		}
		t := float64(v-minV) / float64(maxV-minV)
		return plotY0 + (1-float32(t))*plotH
	}

	// -----------------------
	// Draw lines (one per series)
	// -----------------------
	if n >= 2 {
		for j := 0; j < d.vars; j++ {
			if !d.variableShown(j) {
				continue
			}

			if j >= len(d.values) {
				break
			}
			series := d.values[j]
			if len(series) < n {
				continue
			}
			c := d.seriesColor(j)

			for i := 0; i < n-1; i++ {
				drawLineAA(d.img,
					toX(i), toY(series[i]),
					toX(i+1), toY(series[i+1]),
					c, d.opts.LineStroke,
				)
			}
		}
	}

	// -----------------------
	// Draw markers (series-colored)
	// -----------------------
	if d.opts.ShowMarkers && n > 0 {
		for j := 0; j < d.vars; j++ {
			if !d.variableShown(j) {
				continue
			}

			if j >= len(d.values) {
				break
			}
			series := d.values[j]
			if len(series) < n {
				continue
			}
			c := d.seriesColor(j)

			for i := 0; i < n; i++ {
				drawCircleAA(d.img, toX(i), toY(series[i]), d.opts.MarkerRadius, c)
			}
		}
	}

	// Refresh raster to paint current d.img
	d.raster.Refresh()

	// Update labels
	if d.opts.ShowAxes {
		d.setLabelsLocked(plotX0, plotY0, plotW, plotH)
	} else {
		d.yMinT.Hide()
		d.yMaxT.Hide()
		d.xMinT.Hide()
		d.xMaxT.Hide()
		for i := range d.xTicks {
			d.xTicks[i].Hide()
			d.xTickLines[i].Hide()
		}
	}
}

func (d *LineChartDrawer) setLabelsLocked(plotX0, plotY0, plotW, plotH float32) {
	d.yMinT.Show()
	d.yMaxT.Show()
	d.xMinT.Show()
	d.xMaxT.Show()

	d.yMinT.Color = d.opts.Axis
	d.yMaxT.Color = d.opts.Axis
	d.xMinT.Color = d.opts.Axis
	d.xMaxT.Color = d.opts.Axis

	d.yMinT.Text = fmt.Sprintf("%d", d.minV)
	d.yMaxT.Text = fmt.Sprintf("%d", d.maxV)

	// Place near left side
	d.yMaxT.Move(fyne.NewPos(4, plotY0-2))
	d.yMinT.Move(fyne.NewPos(4, plotY0+plotH-10))
	d.yMaxT.Refresh()
	d.yMinT.Refresh()

	n := d.pointCountSafeLocked()
	d.updateXTickLabelsLocked(plotX0, plotY0, plotW, plotH, n)
}

func (d *LineChartDrawer) updateXTickLabelsLocked(plotX0, plotY0, plotW, plotH float32, n int) {
	// Hide all intermediate ticks by default
	for i := range d.xTicks {
		d.xTicks[i].Hide()
		d.xTickLines[i].Hide()
	}

	if n <= 0 || len(d.times) == 0 {
		d.xMinT.Text, d.xMaxT.Text = "", ""
		d.xMinT.Refresh()
		d.xMaxT.Refresh()
		return
	}

	format := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("15:04:05")
	}

	labelY := plotY0 + plotH + 6
	axisY := plotY0 + plotH

	// Endpoints
	leftT := format(d.times[0])
	rightT := format(d.times[len(d.times)-1])

	d.xMinT.Text = leftT
	d.xMaxT.Text = rightT

	d.xMinT.TextSize = d.xLabelTextSize
	d.xMaxT.TextSize = d.xLabelTextSize

	d.xMinT.Refresh()
	d.xMaxT.Refresh()

	// Proper alignment:
	// Left label at plotX0; Right label right-aligned to plotX0+plotW
	leftSz := d.xMinT.MinSize()
	rightSz := d.xMaxT.MinSize()

	d.xMinT.Move(fyne.NewPos(plotX0, labelY))
	d.xMaxT.Move(fyne.NewPos(plotX0+plotW-rightSz.Width, labelY))

	// Optional: endpoint tick marks on axis
	// (If you don't want them, comment out)
	// We reuse the first two xTickLines for endpoints only when n>1
	if n > 1 && len(d.xTickLines) >= 2 {
		// left endpoint tick at x0
		ln0 := d.xTickLines[0]
		ln0.StrokeColor = d.opts.Axis
		ln0.Position1 = fyne.NewPos(plotX0, axisY)
		ln0.Position2 = fyne.NewPos(plotX0, axisY+4)
		ln0.Show()

		// right endpoint tick at x1
		ln1 := d.xTickLines[1]
		ln1.StrokeColor = d.opts.Axis
		ln1.Position1 = fyne.NewPos(plotX0+plotW, axisY)
		ln1.Position2 = fyne.NewPos(plotX0+plotW, axisY+4)
		ln1.Show()
	}

	_ = leftSz // not used beyond this point; kept if you later want collision checks

	// No intermediate ticks when only one point
	if n == 1 {
		return
	}

	// Density estimation like bar chart
	sampleW := d.sampleTimeLabelWidthLocked()
	minSpacing := sampleW + 8
	maxLabels := int(plotW / minSpacing)

	if maxLabels < 2 {
		maxLabels = 2
	}
	if maxLabels > d.maxXTicks {
		maxLabels = d.maxXTicks
	}
	if maxLabels > n {
		maxLabels = n
	}

	// We count endpoints as labels, so intermediate count is maxLabels-2
	wantMid := maxLabels - 2
	if wantMid <= 0 {
		return
	}

	den := float64(maxLabels - 1) // safe because maxLabels>=2
	midUsed := 0

	// Use remaining slots for mid ticks.
	// IMPORTANT: Because we used xTickLines[0],[1] optionally for endpoints,
	// we start mid tick lines at index 2.
	lineBase := 2
	if len(d.xTickLines) < lineBase+wantMid {
		// if not enough, reduce
		can := len(d.xTickLines) - lineBase
		if can < 0 {
			can = 0
		}
		if wantMid > can {
			wantMid = can
		}
	}
	if wantMid <= 0 {
		return
	}

	for k := 1; k < maxLabels-1; k++ {
		idx := int(math.Round(float64(k) * float64(n-1) / den))
		if idx < 0 {
			idx = 0
		}
		if idx > n-1 {
			idx = n - 1
		}

		cx := plotX0 + (float32(idx)/float32(n-1))*plotW

		// label object
		tlbl := d.xTicks[midUsed]
		tlbl.Color = d.opts.Axis
		tlbl.Text = format(d.times[idx])
		tlbl.TextSize = d.xLabelTextSize
		tlbl.Alignment = fyne.TextAlignCenter
		tlbl.Refresh()

		ls := tlbl.MinSize()
		tlbl.Move(fyne.NewPos(cx-ls.Width/2, labelY))
		tlbl.Show()

		// tick line object
		ln := d.xTickLines[lineBase+midUsed]
		ln.StrokeColor = d.opts.Axis
		ln.Position1 = fyne.NewPos(cx, axisY)
		ln.Position2 = fyne.NewPos(cx, axisY+4)
		ln.Show()

		midUsed++
		if midUsed >= wantMid || midUsed >= len(d.xTicks) {
			break
		}
	}

	// Ensure right label is always visible and not pushed off by cache mismatch
	_ = rightSz
}

/* ---------------------------
   Raster drawing helpers
--------------------------- */

func fillRGBA(img *image.RGBA, c color.Color) {
	r, g, b, a := c.RGBA()
	cr := uint8(r >> 8)
	cg := uint8(g >> 8)
	cb := uint8(b >> 8)
	ca := uint8(a >> 8)

	p := img.Pix
	for i := 0; i < len(p); i += 4 {
		p[i+0] = cr
		p[i+1] = cg
		p[i+2] = cb
		p[i+3] = ca
	}
}

func drawAxes(img *image.RGBA, x0, y0, w, h float32, c color.Color, stroke float32) {
	// Y axis
	drawLineAA(img, x0, y0, x0, y0+h, c, stroke)
	// X axis
	drawLineAA(img, x0, y0+h, x0+w, y0+h, c, stroke)
}

func drawGrid(img *image.RGBA, x0, y0, w, h float32, c color.Color, stroke float32, gx, gy int) {
	if gx > 0 {
		for i := 1; i < gx; i++ {
			x := x0 + (float32(i)/float32(gx))*w
			drawLineAA(img, x, y0, x, y0+h, c, stroke)
		}
	}
	if gy > 0 {
		for j := 1; j < gy; j++ {
			y := y0 + (float32(j)/float32(gy))*h
			drawLineAA(img, x0, y, x0+w, y, c, stroke)
		}
	}
}

// drawLineAA draws an anti-aliased line with a given stroke width.
func drawLineAA(img *image.RGBA, x1, y1, x2, y2 float32, c color.Color, stroke float32) {
	if stroke < 1 {
		stroke = 1
	}

	half := stroke / 2
	steps := int(math.Max(1, float64(stroke)))
	for i := -steps; i <= steps; i++ {
		off := (float32(i) / float32(steps)) * half

		dx := x2 - x1
		dy := y2 - y1
		llen := float32(math.Hypot(float64(dx), float64(dy)))
		if llen == 0 {
			blendPixel(img, int(math.Round(float64(x1))), int(math.Round(float64(y1))), c, 1)
			continue
		}
		nx := -dy / llen
		ny := dx / llen
		ax1 := x1 + nx*off
		ay1 := y1 + ny*off
		ax2 := x2 + nx*off
		ay2 := y2 + ny*off
		wuLine(img, ax1, ay1, ax2, ay2, c)
	}
}

// wuLine draws a single-pixel wide anti-aliased line.
func wuLine(img *image.RGBA, x0, y0, x1, y1 float32, c color.Color) {
	steep := math.Abs(float64(y1-y0)) > math.Abs(float64(x1-x0))
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	gradient := 0.0
	if dx != 0 {
		gradient = dy / dx
	}

	ipart := func(x float64) float64 { return math.Floor(x) }
	fpart := func(x float64) float64 { return x - math.Floor(x) }
	rfpart := func(x float64) float64 { return 1 - fpart(x) }

	xend := math.Round(float64(x0))
	yend := float64(y0) + gradient*(xend-float64(x0))
	xgap := rfpart(float64(x0) + 0.5)
	xpxl1 := int(xend)
	ypxl1 := int(ipart(yend))

	plot := func(x, y int, a float64) {
		if steep {
			blendPixel(img, y, x, c, a)
		} else {
			blendPixel(img, x, y, c, a)
		}
	}

	plot(xpxl1, ypxl1, rfpart(yend)*xgap)
	plot(xpxl1, ypxl1+1, fpart(yend)*xgap)

	intery := yend + gradient

	xend = math.Round(float64(x1))
	yend = float64(y1) + gradient*(xend-float64(x1))
	xgap = fpart(float64(x1) + 0.5)
	xpxl2 := int(xend)
	ypxl2 := int(ipart(yend))

	for x := xpxl1 + 1; x <= xpxl2-1; x++ {
		plot(x, int(ipart(intery)), rfpart(intery))
		plot(x, int(ipart(intery))+1, fpart(intery))
		intery += gradient
	}

	plot(xpxl2, ypxl2, rfpart(yend)*xgap)
	plot(xpxl2, ypxl2+1, fpart(yend)*xgap)
}

func drawCircleAA(img *image.RGBA, cx, cy, r float32, c color.Color) {
	if r <= 0 {
		return
	}
	minX := int(math.Floor(float64(cx - r - 1)))
	maxX := int(math.Ceil(float64(cx + r + 1)))
	minY := int(math.Floor(float64(cy - r - 1)))
	maxY := int(math.Ceil(float64(cy + r + 1)))

	rr := float64(r)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := (float64(x) + 0.5) - float64(cx)
			dy := (float64(y) + 0.5) - float64(cy)
			dist := math.Hypot(dx, dy)
			alpha := 0.0
			if dist <= rr {
				alpha = 1.0
			} else if dist <= rr+1.0 {
				alpha = 1.0 - (dist - rr)
			}
			if alpha > 0 {
				blendPixel(img, x, y, c, alpha)
			}
		}
	}
}

func blendPixel(img *image.RGBA, x, y int, c color.Color, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}

	cr, cg, cb, ca := c.RGBA()
	sr := float64(cr>>8) / 255.0
	sg := float64(cg>>8) / 255.0
	sb := float64(cb>>8) / 255.0
	sa := (float64(ca>>8) / 255.0) * alpha

	i := img.PixOffset(x, y)

	dr := float64(img.Pix[i+0]) / 255.0
	dg := float64(img.Pix[i+1]) / 255.0
	db := float64(img.Pix[i+2]) / 255.0
	da := float64(img.Pix[i+3]) / 255.0

	outA := sa + da*(1-sa)
	if outA <= 0 {
		return
	}
	outR := (sr*sa + dr*da*(1-sa)) / outA
	outG := (sg*sa + dg*da*(1-sa)) / outA
	outB := (sb*sa + db*da*(1-sa)) / outA

	img.Pix[i+0] = uint8(clamp0164(outR) * 255)
	img.Pix[i+1] = uint8(clamp0164(outG) * 255)
	img.Pix[i+2] = uint8(clamp0164(outB) * 255)
	img.Pix[i+3] = uint8(clamp0164(outA) * 255)
}

func clamp0164(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
