package fynerenderer

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type LegendItem struct {
	Label string
	Color color.Color
}

type DiagramHeader struct {
	root *fyne.Container

	title *widget.Label
	last  *widget.Label

	legendBox *fyne.Container // holds rows
	perRow    int             // legend items per row (simple wrapping)
}

func NewDiagramHeader(title string, perRow int) *DiagramHeader {
	if perRow <= 0 {
		perRow = 4
	}

	h := &DiagramHeader{
		title:     widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		last:      widget.NewLabel(""),
		legendBox: container.NewVBox(),
		perRow:    perRow,
	}

	// Title row: [Title .......... LastUpdated]
	titleRow := container.NewHBox(
		h.title,
		layout.NewSpacer(),
		h.last,
	)

	// Root: title row + legend rows
	h.root = container.NewVBox(
		titleRow,
		h.legendBox,
	)

	// Hide last-updated by default (until set)
	h.last.Hide()
	// Hide legend by default (until set)
	h.legendBox.Hide()

	return h
}

func (h *DiagramHeader) Root() fyne.CanvasObject { return h.root }

func (h *DiagramHeader) SetTitle(t string) {
	h.title.SetText(t)
}

func (h *DiagramHeader) SetLastUpdated(at time.Time) {
	h.last.SetText(fmt.Sprintf("Last: %s", at.Format("15:04:05")))
	h.last.Show()
}

func (h *DiagramHeader) ClearLastUpdated() {
	h.last.SetText("")
	h.last.Hide()
}

func (h *DiagramHeader) SetLegend(items []LegendItem) {
	h.legendBox.Objects = h.legendBox.Objects[:0]

	// No legend for 0-1 series
	if len(items) <= 1 {
		h.legendBox.Hide()
		h.legendBox.Refresh()
		return
	}

	// Build rows of legend items (simple wrap by count)
	for i := 0; i < len(items); i += h.perRow {
		end := i + h.perRow
		if end > len(items) {
			end = len(items)
		}
		row := container.NewHBox()
		for _, it := range items[i:end] {
			row.Add(newLegendEntry(it))
			row.Add(layout.NewSpacer()) // gives breathing space; remove if you want tight packing
		}
		h.legendBox.Add(row)
	}

	h.legendBox.Show()
	h.legendBox.Refresh()
}

func newLegendEntry(it LegendItem) fyne.CanvasObject {
	// color swatch
	swatch := canvas.NewRectangle(it.Color)
	swatch.SetMinSize(fyne.NewSize(12, 12))

	// label
	lbl := widget.NewLabel(it.Label)
	lbl.TextStyle = fyne.TextStyle{} // keep normal; tweak if you want smaller/bold

	return container.NewHBox(
		swatch,
		layout.NewSpacer(), // tiny gap; swap for layout.NewSpacer() if you want larger spacing control
		lbl,
	)
}
