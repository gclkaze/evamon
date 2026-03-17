package fynediagrams

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

const (
	ttPadH   = float32(10)
	ttPadV   = float32(6)
	ttLineH  = float32(16)
	ttTextSz = float32(11)
	ttDotGap = float32(4)
	ttMaxRows = 8
)

type tooltipRow struct {
	dot  *canvas.Text // "●" in the series color
	text *canvas.Text // "name:  value"
}

// chartTooltip is a floating overlay positioned near the hovered data point.
// It is owned by the widget and lives inside the renderer's object list so it
// is always drawn on top of the chart raster.
type chartTooltip struct {
	bg        *canvas.Rectangle
	timeLabel *canvas.Text
	rows      []*tooltipRow
	root      *fyne.Container
}

func newChartTooltip() *chartTooltip {
	bg := canvas.NewRectangle(color.NRGBA{R: 18, G: 18, B: 18, A: 215})
	bg.CornerRadius = 6

	tl := canvas.NewText("", color.NRGBA{R: 160, G: 160, B: 160, A: 255})
	tl.TextSize = ttTextSz

	objs := []fyne.CanvasObject{bg, tl}
	rows := make([]*tooltipRow, ttMaxRows)
	for i := 0; i < ttMaxRows; i++ {
		dot := canvas.NewText("●", color.White)
		dot.TextSize = ttTextSz
		txt := canvas.NewText("", color.NRGBA{R: 230, G: 230, B: 230, A: 255})
		txt.TextSize = ttTextSz
		rows[i] = &tooltipRow{dot: dot, text: txt}
		objs = append(objs, dot, txt)
	}

	c := container.NewWithoutLayout(objs...)
	c.Hide()
	return &chartTooltip{bg: bg, timeLabel: tl, rows: rows, root: c}
}

func (t *chartTooltip) Object() fyne.CanvasObject { return t.root }

// Update positions and displays the tooltip near mousePos.
// shownIndices determines which series rows are rendered.
func (t *chartTooltip) Update(
	mousePos fyne.Position,
	widgetSize fyne.Size,
	at time.Time,
	variables []dport.VariableStyle,
	seriesVals []int,
	shownIndices map[int]bool,
) {
	t.timeLabel.Text = at.Format("15:04:05")
	t.timeLabel.Refresh()

	visCount := 0
	for i := 0; i < ttMaxRows; i++ {
		row := t.rows[i]
		show := i < len(variables) && i < len(seriesVals)
		if show && shownIndices != nil {
			if v, ok := shownIndices[i]; ok && !v {
				show = false
			}
		}
		if !show {
			row.dot.Hide()
			row.text.Hide()
			continue
		}
		c := variables[i].VarColor
		if c == nil {
			c = color.White
		}
		row.dot.Color = c
		row.dot.Show()
		row.dot.Refresh()
		row.text.Text = fmt.Sprintf("%s:  %d", variables[i].VariableName, seriesVals[i])
		row.text.Show()
		row.text.Refresh()
		visCount++
	}

	if visCount == 0 {
		t.root.Hide()
		return
	}

	// Measure tooltip width.
	maxContentW := t.timeLabel.MinSize().Width
	for i := 0; i < ttMaxRows; i++ {
		if !t.rows[i].dot.Visible() {
			continue
		}
		rowW := t.rows[i].dot.MinSize().Width + ttDotGap + t.rows[i].text.MinSize().Width
		if rowW > maxContentW {
			maxContentW = rowW
		}
	}
	bgW := maxContentW + ttPadH*2
	bgH := ttPadV*2 + ttLineH + float32(visCount)*ttLineH

	x, y := t.computePos(mousePos, widgetSize, bgW, bgH)

	t.bg.Move(fyne.NewPos(x, y))
	t.bg.Resize(fyne.NewSize(bgW, bgH))
	t.bg.Refresh()

	t.timeLabel.Move(fyne.NewPos(x+ttPadH, y+ttPadV))
	t.timeLabel.Refresh()

	rowY := y + ttPadV + ttLineH
	for i := 0; i < ttMaxRows; i++ {
		row := t.rows[i]
		if !row.dot.Visible() {
			continue
		}
		dotSz := row.dot.MinSize()
		txtSz := row.text.MinSize()
		row.dot.Move(fyne.NewPos(x+ttPadH, rowY+(ttLineH-dotSz.Height)/2))
		row.dot.Refresh()
		row.text.Move(fyne.NewPos(x+ttPadH+dotSz.Width+ttDotGap, rowY+(ttLineH-txtSz.Height)/2))
		row.text.Refresh()
		rowY += ttLineH
	}

	t.root.Show()
	t.root.Refresh()
}

func (t *chartTooltip) computePos(mouse fyne.Position, widget fyne.Size, w, h float32) (x, y float32) {
	const offset = float32(14)
	x = mouse.X + offset
	if x+w > widget.Width-4 {
		x = mouse.X - w - offset
	}
	y = mouse.Y - h/2
	if y < 4 {
		y = 4
	}
	if y+h > widget.Height-4 {
		y = widget.Height - h - 4
	}
	return
}

func (t *chartTooltip) Hide() {
	t.root.Hide()
	t.root.Refresh()
}
