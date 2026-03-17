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

	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	port "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

/*
LineChartDrawer draws multi-series points into a Raster and overlays axis labels.
Multi-series behavior:
  - vars == 1: Push accepts int/float values
  - vars > 1: Push expects a JSON array string (e.g. "[1,2,3]") and unmarshals to []int

X-axis labels:
  - Always shows first and last timestamps (aligned to plot bounds)
  - Adds intermediate tick labels (overlap-aware) with reused objects
*/
type LineChartDrawer struct {
	mu sync.Mutex

	opts port.LineChartOptions

	root   *fyne.Container
	raster *canvas.Raster

	// Data
	data data.IMultiSeriesData
	vars int

	// Scale cache (used by labels)
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
	maxXTicks  int // inclusive count used for density calculation

	// Cached label measurement for overlap-aware density
	xLabelTextSize float32
	xLabelSampleW  float32
	xLabelCacheOK  bool

	// Last known size
	size fyne.Size

	// Raster backing buffer
	img *image.RGBA

	// Series visibility
	shownIndices map[int]bool

	// zoomWindow is the number of most-recent points to display.
	// 0 means show all buffered points (no zoom applied).
	zoomWindow int

	// refreshHook, when set, is called instead of d.raster.Refresh().
	// Used in maximize mode so refreshes go through the maximize adapter
	// (whose canvas is reliably tracked) rather than CanvasForObject(d.raster).
	refreshHook func()
}

// refreshRaster calls the refresh hook if one is installed (maximize mode),
// otherwise falls back to d.raster.Refresh() for normal operation.
func (d *LineChartDrawer) refreshRaster() {
	if d.refreshHook != nil {
		d.refreshHook()
	} else {
		d.raster.Refresh()
	}
}

type plotRect struct {
	x0, y0, w, h float32
}

type frame struct {
	wpx, hpx int
	plot     plotRect
	n        int
	minV     int
	maxV     int
}

type mapper struct {
	plot plotRect
	n    int
	minV int
	maxV int
}

func (m mapper) X(i int) float32 {
	if m.n <= 1 {
		return m.plot.x0 + m.plot.w/2
	}
	return m.plot.x0 + (float32(i)/float32(m.n-1))*m.plot.w
}

func (m mapper) Y(v int) float32 {
	if m.maxV == m.minV {
		return m.plot.y0 + m.plot.h/2
	}
	t := float64(v-m.minV) / float64(m.maxV-m.minV)
	return m.plot.y0 + (1-float32(t))*m.plot.h
}

func NewLineChartDrawer(src data.IMultiSeriesData, opts port.LineChartOptions, initialWidth, initialHeight float32) *LineChartDrawer {
	opts = normalizeLineOpts(opts)

	if initialWidth <= 0 {
		initialWidth = 800
	}
	if initialHeight <= 0 {
		initialHeight = 260
	}

	d := &LineChartDrawer{
		opts:           opts,
		size:           fyne.NewSize(initialWidth, initialHeight),
		xLabelTextSize: 10,
		maxXTicks:      6,
	}

	/*	d.vars = len(opts.Variables)
		if d.vars <= 0 {
			d.vars = 1
		}*/

	d.data = src
	d.vars = 1
	if d.data != nil && d.data.Vars() > 0 {
		d.vars = d.data.Vars()
	} else {
		d.vars = len(opts.Variables)
		if d.vars <= 0 {
			d.vars = 1
		}
	}

	// NOW init shownIndices
	d.shownIndices = make(map[int]bool, d.vars)
	for i := 0; i < d.vars; i++ {
		d.shownIndices[i] = true
	}

	//d.values = make([][]int, d.vars)

	// Raster draws the full frame inside the generator (robust across windows/backends).
	d.raster = canvas.NewRaster(func(w, h int) image.Image {
		d.mu.Lock()
		defer d.mu.Unlock()

		if w <= 1 || h <= 1 {
			return image.NewRGBA(image.Rect(0, 0, 2, 2))
		}

		if d.img == nil || d.img.Bounds().Dx() != w || d.img.Bounds().Dy() != h {
			d.img = image.NewRGBA(image.Rect(0, 0, w, h))
		}

		d.drawLocked(w, h)
		return d.img
	})

	d.initOverlayObjects()

	objects := []fyne.CanvasObject{d.raster, d.yMinT, d.yMaxT, d.xMinT, d.xMaxT}
	for i := 0; i < d.maxXTicks; i++ {
		objects = append(objects, d.xTickLines[i], d.xTicks[i])
	}
	d.root = container.NewWithoutLayout(objects...)
	d.root.Resize(d.size)

	d.raster.Move(fyne.NewPos(0, 0))
	d.raster.Resize(d.size)

	// Initial paint
	d.raster.Refresh()
	return d
}

func normalizeLineOpts(opts port.LineChartOptions) port.LineChartOptions {
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
	// Pads/colors/strokes are assumed to be set by DefaultLineChartOptions,
	// but we keep as-is here to avoid surprising overrides.
	return opts
}

func (d *LineChartDrawer) initOverlayObjects() {
	// Axis labels
	d.yMinT = canvas.NewText("", d.opts.Axis)
	d.yMaxT = canvas.NewText("", d.opts.Axis)
	d.xMinT = canvas.NewText("", d.opts.Axis)
	d.xMaxT = canvas.NewText("", d.opts.Axis)

	d.yMinT.TextSize = d.xLabelTextSize
	d.yMaxT.TextSize = d.xLabelTextSize
	d.xMinT.TextSize = d.xLabelTextSize
	d.xMaxT.TextSize = d.xLabelTextSize

	d.yMinT.Alignment = fyne.TextAlignLeading
	d.yMaxT.Alignment = fyne.TextAlignLeading
	d.xMinT.Alignment = fyne.TextAlignLeading
	d.xMaxT.Alignment = fyne.TextAlignLeading

	// Mid ticks (labels + lines), reused.
	for i := 0; i < d.maxXTicks; i++ {
		t := canvas.NewText("", d.opts.Axis)
		t.TextSize = d.xLabelTextSize
		t.Alignment = fyne.TextAlignCenter
		t.Hide()
		d.xTicks = append(d.xTicks, t)

		ln := canvas.NewLine(d.opts.Axis)
		ln.StrokeWidth = 1
		ln.Hide()
		d.xTickLines = append(d.xTickLines, ln)
	}
}

func (d *LineChartDrawer) Object() fyne.CanvasObject { return d.root }

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

	d.raster.Refresh()
}

func (d *LineChartDrawer) SetOptions(opts port.LineChartOptions) {
	d.mu.Lock()
	defer d.mu.Unlock()

	opts = normalizeLineOpts(opts)

	newVars := len(opts.Variables)
	if newVars <= 0 {
		newVars = 1
	}
	if newVars != d.vars {
		d.vars = newVars
		//d.values = make([][]int, d.vars)
		//d.times = nil
		d.shownIndices = make(map[int]bool, d.vars)
		for i := 0; i < d.vars; i++ {
			d.shownIndices[i] = true
		}
	}

	d.opts = opts

	// Keep max points trimming consistent.
	//d.trimLocked()

	d.refreshRaster()
}

func (d *LineChartDrawer) variableShown(i int) bool { return d.shownIndices[i] }

func (d *LineChartDrawer) ToggleItem(it *uport.LegendItem) {
	if len(d.shownIndices) == 1 {
		return
	}
	d.mu.Lock()
	d.shownIndices[it.Index] = !d.shownIndices[it.Index]
	d.mu.Unlock()

	// repaint
	d.refreshRaster()
}

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
		v, ok := anyToNonNegInt(val)
		if !ok || d.data == nil {
			return
		}

		// append as 1-element slice
		d.data.Append(at, []int{v})
		d.refreshRaster()
	})
}

func (d *LineChartDrawer) handleMultiVariableInput(at time.Time, nums *[]int) {
	UI(func() {
		if d.data == nil || len(*nums) != d.vars {
			return
		}
		d.data.Append(at, *nums)
		d.refreshRaster()
	})
}

func anyToNonNegInt(val any) (int, bool) {
	switch x := val.(type) {
	case int:
		if x < 0 {
			return 0, true
		}
		return x, true
	case float64:
		v := int(x)
		if v < 0 {
			v = 0
		}
		return v, true
	case float32:
		v := int(x)
		if v < 0 {
			v = 0
		}
		return v, true
	default:
		return 0, false
	}
}

func (d *LineChartDrawer) pointCountSafeFromSnapshot(n int, values [][]int) int {
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

func (d *LineChartDrawer) seriesColor(j int) color.Color {
	if j >= 0 && j < len(d.opts.Variables) && d.opts.Variables[j].VarColor != nil {
		return d.opts.Variables[j].VarColor
	}
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
func (d *LineChartDrawer) GetDataSeries() data.IMultiSeriesData {
	return d.data
}

func (d *LineChartDrawer) ZoomIn() {
	d.mu.Lock()
	if d.data == nil {
		d.mu.Unlock()
		return
	}
	current := d.data.Len()
	if d.zoomWindow == 0 {
		if current <= zoomMin {
			d.mu.Unlock()
			return // not enough data to narrow the view
		}
		d.zoomWindow = current - zoomStep
	} else {
		if d.zoomWindow <= zoomMin {
			d.mu.Unlock()
			return // already at minimum zoom
		}
		d.zoomWindow -= zoomStep
	}
	if d.zoomWindow < zoomMin {
		d.zoomWindow = zoomMin
	}
	d.mu.Unlock()
	d.refreshRaster()
}

func (d *LineChartDrawer) ZoomOut() {
	d.mu.Lock()
	if d.zoomWindow == 0 {
		d.mu.Unlock()
		return
	}
	total := 0
	if d.data != nil {
		total = d.data.MaxPoints()
	}
	d.zoomWindow += zoomStep
	if d.zoomWindow >= total {
		d.zoomWindow = 0
	}
	d.mu.Unlock()
	d.refreshRaster()
}

/* ---------------------------
   Rendering (called from raster generator)
--------------------------- */

func (d *LineChartDrawer) zoomedStartLocked() int {
	if d.zoomWindow <= 0 || d.data == nil {
		return 0
	}
	start := d.data.Len() - d.zoomWindow
	if start < 0 {
		return 0
	}
	return start
}

func (d *LineChartDrawer) drawLocked(w, h int) {
	times, values := d.data.ReadWindow(d.zoomedStartLocked(), 0) // copies
	f := d.computeFrameFromSnapshotLocked(w, h, times, values)

	fillRGBA(d.img, d.opts.Background)
	d.drawGridAxesLocked(f.plot)

	m := mapper{plot: f.plot, n: f.n, minV: f.minV, maxV: f.maxV}
	d.drawLinesLocked(f.n, m, values)
	d.drawMarkersLocked(f.n, m, values)

	d.updateOverlayLocked(f.plot, f.n, times)
}

func (d *LineChartDrawer) computeFrameFromSnapshotLocked(w, h int, times []time.Time, values [][]int) frame {
	plot := plotRect{
		x0: d.opts.PadL,
		y0: d.opts.PadT,
		w:  float32(w) - d.opts.PadL - d.opts.PadR,
		h:  float32(h) - d.opts.PadT - d.opts.PadB,
	}
	if plot.w < 10 {
		plot.w = 10
	}
	if plot.h < 10 {
		plot.h = 10
	}

	n := d.pointCountSafeFromSnapshot(len(times), values) // <-- key line

	minV, maxV := d.computeMinMaxFromSnapshotLocked(n, values)
	d.minV, d.maxV = minV, maxV

	return frame{wpx: w, hpx: h, plot: plot, n: n, minV: minV, maxV: maxV}
}

func (d *LineChartDrawer) computeMinMaxFromSnapshotLocked(n int, values [][]int) (minV, maxV int) {
	if n <= 0 {
		return 0, 0
	}

	seeded := false
	for j := 0; j < d.vars && !seeded; j++ {
		if !d.variableShown(j) || j >= len(values) || len(values[j]) < n {
			continue
		}
		minV, maxV = values[j][0], values[j][0]
		seeded = true
	}
	if !seeded {
		return 0, 0
	}

	for j := 0; j < d.vars; j++ {
		if !d.variableShown(j) || j >= len(values) || len(values[j]) < n {
			continue
		}
		series := values[j]
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
	return minV - pad, maxV + pad
}

func (d *LineChartDrawer) drawGridAxesLocked(p plotRect) {
	if d.opts.ShowGrid {
		drawGrid(d.img, p.x0, p.y0, p.w, p.h, d.opts.Grid, d.opts.GridStroke, d.opts.GridX, d.opts.GridY)
	}
	if d.opts.ShowAxes {
		drawAxes(d.img, p.x0, p.y0, p.w, p.h, d.opts.Axis, d.opts.AxisStroke)
	}
}

func (d *LineChartDrawer) drawLinesLocked(n int, m mapper, values [][]int) {
	if n < 2 {
		return
	}
	for j := 0; j < d.vars; j++ {
		if !d.variableShown(j) || j >= len(values) || len(values[j]) < n {
			continue
		}
		series := values[j]
		c := d.seriesColor(j)
		for i := 0; i < n-1; i++ {
			drawLineAA(d.img, m.X(i), m.Y(series[i]), m.X(i+1), m.Y(series[i+1]), c, d.opts.LineStroke)
		}
	}
}

func (d *LineChartDrawer) drawMarkersLocked(n int, m mapper, values [][]int) {
	if !d.opts.ShowMarkers || n <= 0 {
		return
	}
	for j := 0; j < d.vars; j++ {
		if !d.variableShown(j) || j >= len(values) || len(values[j]) < n {
			continue
		}
		series := values[j]
		c := d.seriesColor(j)
		for i := 0; i < n; i++ {
			drawCircleAA(d.img, m.X(i), m.Y(series[i]), d.opts.MarkerRadius, c)
		}
	}
}

func (d *LineChartDrawer) updateOverlayLocked(p plotRect, n int, times []time.Time) {
	if d.opts.ShowAxes {
		d.setLabelsLocked(p, n, times)
		return
	}
	d.hideAllLabelsLocked()
}

/* ---------------------------
   Labels / ticks
--------------------------- */

func (d *LineChartDrawer) setLabelsLocked(p plotRect, n int, times []time.Time) {
	d.showAxisLabelsLocked()
	d.setYLabelsLocked(p)
	d.updateXTickLabelsLocked(p, n, times)
}

func (d *LineChartDrawer) showAxisLabelsLocked() {
	d.yMinT.Show()
	d.yMaxT.Show()
	d.xMinT.Show()
	d.xMaxT.Show()

	d.yMinT.Color = d.opts.Axis
	d.yMaxT.Color = d.opts.Axis
	d.xMinT.Color = d.opts.Axis
	d.xMaxT.Color = d.opts.Axis
}

func (d *LineChartDrawer) setYLabelsLocked(p plotRect) {
	d.yMinT.Text = fmt.Sprintf("%d", d.minV)
	d.yMaxT.Text = fmt.Sprintf("%d", d.maxV)

	// Place near left side
	d.yMaxT.Move(fyne.NewPos(4, p.y0-2))
	d.yMinT.Move(fyne.NewPos(4, p.y0+p.h-10))

	d.yMaxT.Refresh()
	d.yMinT.Refresh()
}

func (d *LineChartDrawer) updateXTickLabelsLocked(p plotRect, n int, times []time.Time) {
	d.hideAllMidTicksLocked()

	if n <= 0 || len(times) == 0 {
		d.xMinT.Text, d.xMaxT.Text = "", ""
		d.xMinT.Refresh()
		d.xMaxT.Refresh()
		return
	}

	axisY := p.y0 + p.h
	labelY := axisY + 6

	// Endpoints
	d.setXEndpointsLocked(p, labelY, axisY, times)

	if n == 1 {
		return
	}

	// Intermediate ticks
	d.setXMidTicksLocked(p, n, labelY, axisY, times)
}

func (d *LineChartDrawer) setXEndpointsLocked(p plotRect, labelY, axisY float32, times []time.Time) {
	format := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("15:04:05")
	}

	leftT := format(times[0])
	rightT := format(times[len(times)-1])

	d.xMinT.Text = leftT
	d.xMaxT.Text = rightT

	d.xMinT.TextSize = d.xLabelTextSize
	d.xMaxT.TextSize = d.xLabelTextSize

	d.xMinT.Refresh()
	d.xMaxT.Refresh()

	rightSz := d.xMaxT.MinSize()

	d.xMinT.Move(fyne.NewPos(p.x0, labelY))
	d.xMaxT.Move(fyne.NewPos(p.x0+p.w-rightSz.Width, labelY))

	// Endpoint tick marks (reuse xTickLines[0],[1])
	if len(d.xTickLines) >= 2 {
		ln0 := d.xTickLines[0]
		ln0.StrokeColor = d.opts.Axis
		ln0.Position1 = fyne.NewPos(p.x0, axisY)
		ln0.Position2 = fyne.NewPos(p.x0, axisY+4)
		ln0.Show()

		ln1 := d.xTickLines[1]
		ln1.StrokeColor = d.opts.Axis
		ln1.Position1 = fyne.NewPos(p.x0+p.w, axisY)
		ln1.Position2 = fyne.NewPos(p.x0+p.w, axisY+4)
		ln1.Show()
	}
}

func (d *LineChartDrawer) setXMidTicksLocked(p plotRect, n int, labelY, axisY float32, times []time.Time) {
	// overlap-aware density
	sampleW := d.sampleTimeLabelWidthLocked()
	minSpacing := sampleW + 8
	maxLabels := int(p.w / minSpacing)

	if maxLabels < 2 {
		maxLabels = 2
	}
	if maxLabels > d.maxXTicks {
		maxLabels = d.maxXTicks
	}
	if maxLabels > n {
		maxLabels = n
	}

	wantMid := maxLabels - 2
	if wantMid <= 0 {
		return
	}

	// We consumed xTickLines[0],[1] for endpoints, so mid tick lines start at 2.
	lineBase := 2
	canLines := len(d.xTickLines) - lineBase
	if canLines < 0 {
		canLines = 0
	}
	if wantMid > canLines {
		wantMid = canLines
	}
	if wantMid > len(d.xTicks) {
		wantMid = len(d.xTicks)
	}
	if wantMid <= 0 {
		return
	}

	format := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("15:04:05")
	}

	den := float64(maxLabels - 1) // safe (maxLabels>=2)
	midUsed := 0

	for k := 1; k < maxLabels-1; k++ {
		idx := int(math.Round(float64(k) * float64(n-1) / den))
		if idx < 0 {
			idx = 0
		}
		if idx > n-1 {
			idx = n - 1
		}

		cx := p.x0 + (float32(idx)/float32(n-1))*p.w

		// label
		tlbl := d.xTicks[midUsed]
		tlbl.Color = d.opts.Axis
		tlbl.Text = format(times[idx])
		tlbl.TextSize = d.xLabelTextSize
		tlbl.Alignment = fyne.TextAlignCenter
		tlbl.Refresh()

		ls := tlbl.MinSize()
		tlbl.Move(fyne.NewPos(cx-ls.Width/2, labelY))
		tlbl.Show()

		// tick line
		ln := d.xTickLines[lineBase+midUsed]
		ln.StrokeColor = d.opts.Axis
		ln.Position1 = fyne.NewPos(cx, axisY)
		ln.Position2 = fyne.NewPos(cx, axisY+4)
		ln.Show()

		midUsed++
		if midUsed >= wantMid {
			break
		}
	}
}

func (d *LineChartDrawer) hideAllMidTicksLocked() {
	for i := range d.xTicks {
		d.xTicks[i].Hide()
	}
	for i := range d.xTickLines {
		d.xTickLines[i].Hide()
	}
}

func (d *LineChartDrawer) hideAllLabelsLocked() {
	d.yMinT.Hide()
	d.yMaxT.Hide()
	d.xMinT.Hide()
	d.xMaxT.Hide()
	d.hideAllMidTicksLocked()
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
	drawLineAA(img, x0, y0, x0, y0+h, c, stroke)     // Y
	drawLineAA(img, x0, y0+h, x0+w, y0+h, c, stroke) // X
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

// Vars returns the number of data series configured on this drawer.
func (d *LineChartDrawer) Vars() int { return d.vars }

// Variables returns the per-series style information (name, color).
func (d *LineChartDrawer) Variables() []port.VariableStyle { return d.opts.Variables }

// ShownIndices returns a copy of the current series-visibility map.
func (d *LineChartDrawer) ShownIndices() map[int]bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make(map[int]bool, len(d.shownIndices))
	for k, v := range d.shownIndices {
		out[k] = v
	}
	return out
}

// HitTestX maps a logical-unit X position (Fyne dp) to the nearest data point
// index in the current zoom window. widgetW is the full widget width in dp.
// Returns (idx, times, values, true) on success, or (0, nil, nil, false) when
// there is no data.
func (d *LineChartDrawer) HitTestX(pixX, widgetW float32) (idx int, times []time.Time, values [][]int, found bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.data == nil {
		return
	}

	plotX0 := d.opts.PadL
	plotW := widgetW - d.opts.PadL - d.opts.PadR
	if plotW <= 1 {
		return
	}

	start := d.zoomedStartLocked()
	times, values = d.data.ReadWindow(start, 0)
	n := d.pointCountSafeFromSnapshot(len(times), values)
	if n == 0 {
		return
	}
	if n < len(times) {
		times = times[:n]
	}

	// Clamp to plot bounds so edge values are reachable.
	if pixX < plotX0 {
		pixX = plotX0
	} else if pixX > plotX0+plotW {
		pixX = plotX0 + plotW
	}

	t := (pixX - plotX0) / plotW
	idx = int(math.Round(float64(t) * float64(n-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return idx, times, values, true
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
