package fynediagrams

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	"github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// BarChartDrawer draws the last N integer values as vertical bars.
type BarChartDrawer struct {
	maxPoints int
	width     float32
	height    float32

	root *fyne.Container

	data      data.IMultiSeriesData
	variables []port.VariableStyle

	bgColor color.Color

	bg []*canvas.Rectangle

	vars int

	xLabelTextSize float32
	xLabelSampleW  float32
	xLabelCacheOK  bool

	shownIndices map[int]bool
}

func (d *BarChartDrawer) Redraw() {
	d.redrawWithAxis()
}

func (d *BarChartDrawer) GetDataSeries() data.IMultiSeriesData {
	return d.data
}

func NewBarChartDrawer(src data.IMultiSeriesData, width, height float32, variables []port.VariableStyle, bgColor color.Color) *BarChartDrawer {
	if width <= 0 {
		width = 600
	}
	if height <= 0 {
		height = 240
	}

	d := &BarChartDrawer{
		width:          width,
		height:         height,
		bgColor:        bgColor,
		xLabelTextSize: 11,
		variables:      variables,
		data:           src,
	}

	// vars: prefer data source
	d.vars = 1
	if d.data != nil && d.data.Vars() > 0 {
		d.vars = d.data.Vars()
	} else {
		d.vars = len(variables)
		if d.vars <= 0 {
			d.vars = 1
		}
	}

	// shownIndices AFTER vars
	d.shownIndices = make(map[int]bool, d.vars)
	for i := 0; i < d.vars; i++ {
		d.shownIndices[i] = true
	}

	d.root = container.NewWithoutLayout()
	d.root.Resize(fyne.NewSize(width, height))

	d.initRectangles(d.vars)

	return d
}

func (d *BarChartDrawer) initRectangles(l int) {
	d.bg = make([]*canvas.Rectangle, 0, l)

	for i := 0; i < l; i++ {
		p := canvas.NewRectangle(d.bgColor)
		d.bg = append(d.bg, p)
		p.Move(fyne.NewPos(0, 0))
		d.root.Add(p)
	}
}

func (d *BarChartDrawer) InvalidateLabelCache() {
	d.xLabelCacheOK = false
	d.xLabelSampleW = 0
}

func (d *BarChartDrawer) SetLabelTextSize(size float32) {
	if size <= 0 || d.xLabelTextSize == size {
		return
	}
	d.xLabelTextSize = size
	d.InvalidateLabelCache()
	d.Redraw()
}

func (d *BarChartDrawer) addAllRects() {

	for i := range d.bg {
		if d.variableShown(i) {
			d.root.Add(d.bg[i])
		}
	}
}

func (d *BarChartDrawer) resizeAllRects(w, h float32) {
	for i := range d.bg {
		if d.variableShown(i) {
			d.bg[i].Resize(fyne.NewSize(w, h))
		}
	}
}

func (d *BarChartDrawer) Root() fyne.CanvasObject { return d.root }

func (d *BarChartDrawer) Push(at time.Time, val any) {
	if d.vars == 1 {
		d.handleMonoVariableInput(at, val)
		return
	}
	slice, ok := val.(string)
	if !ok {
		return
	}
	var nums []int
	if err := json.Unmarshal([]byte(slice), &nums); err != nil {
		return
	}
	d.handleMultiVariableInput(at, &nums)
}

func (d *BarChartDrawer) handleMonoVariableInput(at time.Time, val any) {
	UI(func() {
		if d.data == nil {
			return
		}
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

		d.data.Append(at, []int{v})
		d.redrawWithAxis()
	})
}
func (d *BarChartDrawer) handleMultiVariableInput(at time.Time, nums *[]int) {
	UI(func() {
		if d.data == nil || len(*nums) != d.vars {
			return
		}
		d.data.Append(at, *nums)
		d.redrawWithAxis()
	})
}

func (d *BarChartDrawer) getMaxValueFromSnapshot(values [][]int, n int) int {
	maxV := 0
	for j := 0; j < d.vars; j++ {
		if !d.variableShown(j) || j >= len(values) || len(values[j]) < n {
			continue
		}
		for i := 0; i < n; i++ {
			if values[j][i] > maxV {
				maxV = values[j][i]
			}
		}
	}
	return maxV
}
func (d *BarChartDrawer) variableShown(i int) bool {
	if d.shownIndices == nil {
		return true
	}
	v, ok := d.shownIndices[i]
	if !ok {
		return true
	}
	return v
}

func (d *BarChartDrawer) ToggleItem(it *uport.LegendItem) {
	if d.vars <= 1 {
		return
	}
	d.shownIndices[it.Index] = !d.shownIndices[it.Index]

	d.Redraw()
}

func (d *BarChartDrawer) drawRectsFromSnapshot(plotY1, plotH, plotX0, slotW float32, yMin, yMax float64, values [][]int, pointCount int) {
	if d.vars == 0 {
		return
	}

	// Build list of visible series indices
	visible := make([]int, 0, d.vars)
	for j := 0; j < d.vars; j++ {
		if d.variableShown(j) {
			visible = append(visible, j)
		}
	}
	visCount := len(visible)
	if visCount == 0 {
		return
	}

	// Point count must be safe for what we're drawing (visible series + times)
	//pointCount := d.pointCountSafeFromSnapshot(len(times), values)
	if pointCount == 0 {
		return
	}

	groupW := slotW
	innerSlotW := groupW / float32(visCount)
	barW := innerSlotW * 0.8
	if barW < 1 {
		barW = 1
	}

	for i := 0; i < pointCount; i++ {
		for visIdx, j := range visible {
			// Safety: ensure we can read this point for this series
			if j >= len(values) || i >= len(values[j]) {
				continue
			}

			val := float64(values[j][i])
			frac := float32((val - yMin) / (yMax - yMin))
			frac = clamp01(frac)

			bh := frac * plotH
			if bh < 1 {
				bh = 1
			}

			x := plotX0 +
				float32(i)*groupW +
				float32(visIdx)*innerSlotW +
				(innerSlotW-barW)/2

			y := plotY1 - bh

			r := canvas.NewRectangle(d.variables[j].VarColor)
			r.Move(fyne.NewPos(x, y))
			r.Resize(fyne.NewSize(barW, bh))
			d.root.Add(r)
		}
	}
}

// --- small structs to keep signatures clean ---

type chartStyle struct {
	axisColor  color.Color
	gridColor  color.Color
	labelColor color.Color

	marginLeft   float32
	marginRight  float32
	marginTop    float32
	marginBottom float32

	tickLen float32
}

type plotArea struct {
	x0, y0 float32
	x1, y1 float32
	w, h   float32
}

type yScale struct {
	min  float64
	max  float64
	tcks TickSpec
}

// If your computeNiceTicks returns a concrete type, use it instead.
// This is only to make the snippet compile-ish.
type niceTicks struct {
	Min   float64
	Max   float64
	Ticks []float64
}

// --- refactored entrypoint ---

func (d *BarChartDrawer) redrawWithAxis() {
	// 1) Size
	w, h, ok := d.readRootSize()
	if !ok {
		return
	}

	// 2) Keep background, clear dynamic objects
	d.prepareBackground(w, h)

	// 3) Data snapshot
	if d.data == nil {
		d.root.Refresh()
		return
	}
	times, values := d.data.ReadWindow(0, 0) // copies

	// 4) Safe point count
	n := d.pointCountSafeFromSnapshot(len(times), values)
	if n <= 0 {
		d.root.Refresh()
		return
	}

	// Make sure times matches n (defensive)
	if n < len(times) {
		times = times[:n]
	}

	// 5) Style + plot area
	style := d.defaultChartStyle()
	plot, ok := d.computePlotArea(w, h, style)
	if !ok {
		d.root.Refresh()
		return
	}

	// 6) Y scale (based on visible series)
	ys := d.computeYScaleFromSnapshot(6, values, n)

	// 7) Axes + Y grid/labels
	d.drawAxes(plot, style)
	d.drawYGridAndLabels(plot, style, ys)

	// 8) Bars
	slotW := plot.w / float32(n)
	if slotW < 1 {
		slotW = 1
	}
	d.drawRectsFromSnapshot(plot.y1, plot.h, plot.x0, slotW, ys.min, ys.max, values, n)

	// 9) X labels
	d.drawXLabelsFromTimes(plot, style, slotW, times)

	// 10) Refresh
	d.root.Refresh()
}

// --- helpers ---

func (d *BarChartDrawer) readRootSize() (w, h float32, ok bool) {
	sz := d.root.Size()
	w, h = sz.Width, sz.Height
	if w <= 1 || h <= 1 {
		return 0, 0, false
	}
	return w, h, true
}

func (d *BarChartDrawer) prepareBackground(w, h float32) {
	d.resizeAllRects(w, h)

	// Clear everything but keep background.
	d.root.Objects = d.root.Objects[:0]
	d.addAllRects()
}

func (d *BarChartDrawer) defaultChartStyle() chartStyle {
	return chartStyle{
		axisColor:  color.NRGBA{R: 170, G: 170, B: 170, A: 255},
		gridColor:  color.NRGBA{R: 255, G: 255, B: 255, A: 28}, // subtle
		labelColor: color.NRGBA{R: 225, G: 225, B: 225, A: 255},

		// IMPORTANT: leave enough room for y labels (k/M/B) and tick marks.
		marginLeft:   78,
		marginRight:  12,
		marginTop:    12,
		marginBottom: 34,

		tickLen: 6,
	}
}

func (d *BarChartDrawer) computePlotArea(w, h float32, s chartStyle) (plotArea, bool) {
	plot := plotArea{
		x0: s.marginLeft,
		y0: s.marginTop,
		x1: w - s.marginRight,
		y1: h - s.marginBottom,
	}
	plot.w = plot.x1 - plot.x0
	plot.h = plot.y1 - plot.y0

	if plot.w <= 2 || plot.h <= 2 {
		return plotArea{}, false
	}
	return plot, true
}

func (d *BarChartDrawer) computeYScaleFromSnapshot(targetTicks int, values [][]int, n int) yScale {
	maxV := d.getMaxValueFromSnapshot(values, n)
	if maxV <= 0 {
		maxV = 1
	}

	// add ~10% headroom so top bar doesn’t touch ceiling
	maxWithHeadroom := float64(maxV) * 1.10
	t := computeNiceTicks(0, maxWithHeadroom, targetTicks)

	yMin := t.Min
	yMax := t.Max
	if yMax <= yMin {
		yMax = yMin + 1
	}

	// If computeNiceTicks returns your own type, adjust the field names here.
	return yScale{
		min:  yMin,
		max:  yMax,
		tcks: t,
	}
}

func (d *BarChartDrawer) drawAxes(p plotArea, s chartStyle) {
	yAxis := canvas.NewLine(s.axisColor)
	yAxis.Position1 = fyne.NewPos(p.x0, p.y0)
	yAxis.Position2 = fyne.NewPos(p.x0, p.y1)
	d.root.Add(yAxis)

	xAxis := canvas.NewLine(s.axisColor)
	xAxis.Position1 = fyne.NewPos(p.x0, p.y1)
	xAxis.Position2 = fyne.NewPos(p.x1, p.y1)
	d.root.Add(xAxis)
}

func (d *BarChartDrawer) drawYGridAndLabels(p plotArea, s chartStyle, ys yScale) {
	for _, tv := range ys.tcks.Ticks {
		frac := float32((tv - ys.min) / (ys.max - ys.min)) // 0..1
		frac = clamp01(frac)
		y := p.y1 - frac*p.h

		d.drawHorizontalGridLine(p, s, y)
		d.drawYTick(p, s, y)
		d.drawYLabel(p, s, y, tv)
	}
}

func (d *BarChartDrawer) drawHorizontalGridLine(p plotArea, s chartStyle, y float32) {
	grid := canvas.NewLine(s.gridColor)
	grid.Position1 = fyne.NewPos(p.x0, y)
	grid.Position2 = fyne.NewPos(p.x1, y)
	d.root.Add(grid)
}

func (d *BarChartDrawer) drawYTick(p plotArea, s chartStyle, y float32) {
	tick := canvas.NewLine(s.axisColor)
	tick.Position1 = fyne.NewPos(p.x0-s.tickLen, y)
	tick.Position2 = fyne.NewPos(p.x0, y)
	d.root.Add(tick)
}

func (d *BarChartDrawer) drawYLabel(p plotArea, s chartStyle, y float32, tv float64) {
	lbl := canvas.NewText(formatCompact(tv), s.labelColor)
	lbl.TextSize = d.xLabelTextSize
	lbl.Alignment = fyne.TextAlignTrailing
	lbl.Refresh()

	ls := lbl.MinSize()
	lbl.Move(fyne.NewPos(p.x0-s.tickLen-6-ls.Width, y-ls.Height/2))
	d.root.Add(lbl)
}

func (d *BarChartDrawer) drawXLabelsFromTimes(p plotArea, s chartStyle, slotW float32, times []time.Time) {
	pointCount := len(times)
	if pointCount == 0 {
		return
	}

	if pointCount == 1 {
		cx := p.x0 + p.w/2
		d.drawSingleXLabel(cx, p.y1, s, times[0])
		return
	}

	maxLabels := d.computeMaxXLabels(p.w, s.labelColor)
	if maxLabels > pointCount {
		maxLabels = pointCount
	}
	if maxLabels < 2 {
		maxLabels = 2
	}

	den := float64(maxLabels - 1)
	for k := 0; k < maxLabels; k++ {
		idx := int(math.Round(float64(k) * float64(pointCount-1) / den))
		if idx < 0 {
			idx = 0
		}
		if idx > pointCount-1 {
			idx = pointCount - 1
		}

		cx := p.x0 + (float32(idx)+0.5)*slotW
		d.drawSingleXLabel(cx, p.y1, s, times[idx])
	}
}

func (d *BarChartDrawer) computeMaxXLabels(plotW float32, labelColor color.Color) int {
	sampleW := d.getXLabelSampleWidth(labelColor)

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
	return maxLabels
}

func (d *BarChartDrawer) getXLabelSampleWidth(labelColor color.Color) float32 {
	// If you ever change d.xLabelTextSize dynamically, you can extend this cache
	// to invalidate when it changes. As-is, it's constant (11), so this is enough.
	if d.xLabelCacheOK && d.xLabelSampleW > 0 {
		return d.xLabelSampleW
	}

	sample := canvas.NewText("88:88:88", labelColor) // widest-ish time pattern
	sample.TextSize = d.xLabelTextSize
	sample.Alignment = fyne.TextAlignCenter
	sample.Refresh()

	d.xLabelSampleW = sample.MinSize().Width
	d.xLabelCacheOK = true

	// Fallback safety
	if d.xLabelSampleW <= 0 {
		d.xLabelSampleW = 50
	}
	return d.xLabelSampleW
}
func (d *BarChartDrawer) drawSingleXLabel(cx, axisY float32, s chartStyle, t time.Time) {
	lbl := canvas.NewText(t.Format("15:04:05"), s.labelColor)
	lbl.TextSize = d.xLabelTextSize
	lbl.Alignment = fyne.TextAlignCenter
	lbl.Refresh()

	ls := lbl.MinSize()
	lbl.Move(fyne.NewPos(cx-ls.Width/2, axisY+6))
	d.root.Add(lbl)

	// optional: small vertical tick on x-axis
	xTick := canvas.NewLine(s.axisColor)
	xTick.Position1 = fyne.NewPos(cx, axisY)
	xTick.Position2 = fyne.NewPos(cx, axisY+4)
	d.root.Add(xTick)
}
func (d *BarChartDrawer) pointCountSafeFromSnapshot(n int, values [][]int) int {
	if n <= 0 {
		return 0
	}
	pc := n
	for j := 0; j < d.vars; j++ {
		if !d.variableShown(j) {
			continue
		}
		if j >= len(values) {
			return 0
		}
		if len(values[j]) < pc {
			pc = len(values[j])
		}
	}
	if pc < 0 {
		return 0
	}
	return pc
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
