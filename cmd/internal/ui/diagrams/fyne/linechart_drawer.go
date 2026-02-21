package fynediagrams

import (
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
)

// LineChartDrawer draws points into a Raster and overlays axis labels.
type LineChartDrawer struct {
	mu sync.Mutex

	opts port.LineChartOptions

	root *fyne.Container

	// Raster plot (anti-aliased drawing happens here)
	raster *canvas.Raster

	// Data
	values []int
	times  []time.Time

	// Cached scale
	minV int
	maxV int

	// Overlay labels
	yMinT *canvas.Text
	yMaxT *canvas.Text
	xMinT *canvas.Text
	xMaxT *canvas.Text

	// Last known size
	size fyne.Size

	img *image.RGBA
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
		opts: opts,
		size: fyne.NewSize(initialWidth, initialHeight),
	}

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

	d.yMinT.TextSize = 10
	d.yMaxT.TextSize = 10
	d.xMinT.TextSize = 10
	d.xMaxT.TextSize = 10

	// Root with absolute positioning
	d.root = container.NewWithoutLayout(d.raster, d.yMinT, d.yMaxT, d.xMinT, d.xMaxT)
	d.root.Resize(d.size)

	d.raster.Move(fyne.NewPos(0, 0))
	d.raster.Resize(d.size)

	// Initial draw
	d.refreshLocked()

	return d
}
func (d *LineChartDrawer) Object() fyne.CanvasObject {
	return d.root
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
	d.opts = opts

	// Trim if needed
	if len(d.values) > d.opts.MaxPoints {
		over := len(d.values) - d.opts.MaxPoints
		d.values = d.values[over:]
		d.times = d.times[over:]
	}

	d.refreshLocked()
}

// Push adds a point and redraws.
func (d *LineChartDrawer) Push(at time.Time, val any) {
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

		d.values = append(d.values, v)
		d.times = append(d.times, at)
		fmt.Printf("unsupported val type: %T\n", val)
		fmt.Printf("unsupported val type: %T\n", v)
		if len(d.values) > d.opts.MaxPoints {
			over := len(d.values) - d.opts.MaxPoints
			d.values = d.values[over:]
			d.times = d.times[over:]
		}

		d.refreshLocked()
	})

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

	// Scale data
	minV, maxV := 0, 0
	n := len(d.values)

	if n > 0 {
		minV, maxV = d.values[0], d.values[0]
		for i := 1; i < n; i++ {
			if d.values[i] < minV {
				minV = d.values[i]
			}
			if d.values[i] > maxV {
				maxV = d.values[i]
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

	d.minV, d.maxV = minV, maxV

	// Plot mapping
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

	// Draw line
	if n >= 2 {
		for i := 0; i < n-1; i++ {
			drawLineAA(d.img,
				toX(i), toY(d.values[i]),
				toX(i+1), toY(d.values[i+1]),
				d.opts.Line, d.opts.LineStroke,
			)
		}
	}

	// Draw markers
	if d.opts.ShowMarkers && n > 0 {
		for i := 0; i < n; i++ {
			drawCircleAA(d.img, toX(i), toY(d.values[i]), d.opts.MarkerRadius, d.opts.Marker)
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

	// X labels use time range
	format := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("15:04:05")
	}
	if len(d.times) > 0 {
		d.xMinT.Text = format(d.times[0])
		d.xMaxT.Text = format(d.times[len(d.times)-1])
	} else {
		d.xMinT.Text = ""
		d.xMaxT.Text = ""
	}

	d.xMinT.Move(fyne.NewPos(plotX0, plotY0+plotH+6))
	// crude right align
	d.xMaxT.Move(fyne.NewPos(plotX0+plotW-70, plotY0+plotH+6))
	d.xMinT.Refresh()
	d.xMaxT.Refresh()
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
// This is a pragmatic AA: sample coverage around the ideal line with alpha blending.
func drawLineAA(img *image.RGBA, x1, y1, x2, y2 float32, c color.Color, stroke float32) {
	if stroke < 1 {
		stroke = 1
	}

	// For thicker lines, draw multiple offset AA lines
	half := stroke / 2
	steps := int(math.Max(1, float64(stroke)))
	for i := -steps; i <= steps; i++ {
		off := (float32(i) / float32(steps)) * half
		// offset perpendicular
		dx := x2 - x1
		dy := y2 - y1
		len := float32(math.Hypot(float64(dx), float64(dy)))
		if len == 0 {
			blendPixel(img, int(math.Round(float64(x1))), int(math.Round(float64(y1))), c, 1)
			continue
		}
		nx := -dy / len
		ny := dx / len
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

	// helper
	ipart := func(x float64) float64 { return math.Floor(x) }
	fpart := func(x float64) float64 { return x - math.Floor(x) }
	rfpart := func(x float64) float64 { return 1 - fpart(x) }

	// first endpoint
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

	// second endpoint
	xend = math.Round(float64(x1))
	yend = float64(y1) + gradient*(xend-float64(x1))
	xgap = fpart(float64(x1) + 0.5)
	xpxl2 := int(xend)
	ypxl2 := int(ipart(yend))

	// main loop
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
	// simple soft circle: for each pixel in bounding box compute distance and blend
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
			// soft edge: 1px feather
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

	// standard "source over" blend
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
