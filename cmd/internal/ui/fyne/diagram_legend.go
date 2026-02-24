package fynerenderer

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynelayout "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"golang.org/x/image/colornames"

	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// newDiagramLegend renders a legend (color + name per variable).
// If len(items) <= 1 it returns an empty (hidden) container.
// If onClick is nil -> items are non-interactive labels.
// If onClick is non-nil -> items are clickable "label-like buttons" that call onClick(key).
func newDiagramLegend(items []port.LegendItem, perRow int, onClick func(item *port.LegendItem)) fyne.CanvasObject {
	if perRow <= 0 {
		perRow = 6
	}

	root := container.NewVBox()

	/*	if len(items) <= 1 {
		root.Hide()
		return root
	}*/

	// Build rows
	addFilter := len(items) > 1
	for i := 0; i < len(items); i += perRow {
		end := i + perRow
		if end > len(items) {
			end = len(items)
		}

		row := container.NewHBox()

		for _, it := range items[i:end] {
			row.Add(legendEntry(it, onClick, addFilter))
			// fixed-ish gap between entries (spacer expands; keep small by using a tiny empty canvas object)
			row.Add(fixedHGap(5))
		}

		root.Add(row)
	}

	return root
}

func legendEntry(it port.LegendItem, onClick func(it *port.LegendItem), addFilter bool) fyne.CanvasObject {
	size := float32(12)

	c := parseColor(it.Color)
	swatchBox, swatchInner := makeSwatchWithInner(c, size)

	labelText := strings.TrimSpace(it.Label)
	if labelText == "" {
		labelText = strings.TrimSpace(it.Key)
	}
	if labelText == "" {
		labelText = "?"
	}

	active := true
	inactiveText := theme.DisabledColor()

	txt := canvas.NewText(labelText, theme.ForegroundColor())
	txt.TextSize = 12
	txt.Alignment = fyne.TextAlignLeading
	txt.Resize(txt.MinSize())

	overlay := NewTapOverlay(func() {
		if !addFilter {
			return
		}
		active = !active
		if active {
			txt.Color = theme.ForegroundColor()
			//	setSwatchAlpha(swatchInner, 255)
		} else {
			txt.Color = inactiveText
			//	setSwatchAlpha(swatchInner, 110)
		}
		txt.Refresh()
		swatchInner.Refresh()

		if onClick != nil {
			item := it
			onClick(&item)
		}
	})

	// Max overlays the transparent tappable widget on top of the text
	clickableText := container.NewMax(txt, overlay)

	// Important: give overlay same size as text
	overlay.Resize(txt.MinSize())

	return container.NewHBox(
		container.NewCenter(swatchBox),
		fixedHGap(3),
		clickableText,
	)
}

func makeSwatchWithInner(c color.Color, size float32) (fyne.CanvasObject, *canvas.Rectangle) {
	border := canvas.NewRectangle(color.NRGBA{R: 160, G: 160, B: 160, A: 255})
	border.CornerRadius = 2

	inner := canvas.NewRectangle(c)
	inner.CornerRadius = 2

	padding := float32(1)

	content := container.NewWithoutLayout(border, inner)
	border.Resize(fyne.NewSize(size, size))
	inner.Resize(fyne.NewSize(size-2*padding, size-2*padding))
	inner.Move(fyne.NewPos(padding, padding))

	box := container.NewGridWrap(fyne.NewSize(size, size), content)
	return box, inner
}

func setSwatchAlpha(r *canvas.Rectangle, a uint8) {
	// r.FillColor might not be NRGBA; convert safely
	rc, gc, bc, _ := r.FillColor.RGBA()
	r.FillColor = color.NRGBA{
		R: uint8(rc >> 8),
		G: uint8(gc >> 8),
		B: uint8(bc >> 8),
		A: a,
	}
}

func fixedHGap(px float32) fyne.CanvasObject {
	// A tiny transparent rectangle acts like fixed spacing.
	r := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	r.SetMinSize(fyne.NewSize(px, 1))
	return r
}

// parseColor supports "#RRGGBB", "#RRGGBBAA", and a few CSS-like names.
// Unknown values fall back to theme.ForegroundColor() to stay readable.
func parseColor(s string) color.Color {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return theme.ForegroundColor()
	}

	// Hex formats: #RRGGBB or #RRGGBBAA
	if strings.HasPrefix(s, "#") {
		h := strings.TrimPrefix(s, "#")
		if len(h) == 6 || len(h) == 8 {
			r, err1 := strconv.ParseUint(h[0:2], 16, 8)
			g, err2 := strconv.ParseUint(h[2:4], 16, 8)
			b, err3 := strconv.ParseUint(h[4:6], 16, 8)
			if err1 == nil && err2 == nil && err3 == nil {
				a := uint64(255)
				if len(h) == 8 {
					if aa, err := strconv.ParseUint(h[6:8], 16, 8); err == nil {
						a = aa
					}
				}
				return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
			}
		}
	}

	if c, ok := colornames.Map[s]; ok {
		return c
	}
	return theme.ForegroundColor()
}

// Avoid unused import warning if you remove fynelayout later; kept here as a hint.
// (Not actually used; you can delete this line if you wish.)
var _ = fynelayout.NewSpacer
