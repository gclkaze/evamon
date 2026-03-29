package executionlog

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/gclkaze/evamon/cmd/internal/models"
)

// LogLineRenderer builds Fyne canvas objects for individual log lines.
// All colour decisions come from LogStyleConfig; no Fyne theme is consulted.
type LogLineRenderer struct {
	config *LogStyleConfig
}

// NewLogLineRenderer creates a renderer backed by the given config.
func NewLogLineRenderer(config *LogStyleConfig) *LogLineRenderer {
	return &LogLineRenderer{config: config}
}

// Render returns the canvas object for one ExecutionLogEntry.
// Only entry.ParsedMsg is displayed; entry.Stream and entry.Line are ignored.
func (r *LogLineRenderer) Render(entry models.ExecutionLogEntry) fyne.CanvasObject {
	parsed := ParseLogLine(entry.ParsedMsg)
	if !parsed.IsStructured {
		return r.renderFreeForm(entry.ParsedMsg)
	}
	return r.renderStructured(entry.ParsedMsg, parsed.OperationType)
}

func (r *LogLineRenderer) renderStructured(msg string, opType OperationType) fyne.CanvasObject {
	style, ok := r.config.StyleFor(opType)
	if !ok {
		return r.renderFreeForm(msg)
	}
	return r.buildRenderedLine(msg, parseHexColor(style.BorderColor), parseHexColor(style.BadgeColor), style.BadgeText)
}

func (r *LogLineRenderer) renderFreeForm(msg string) fyne.CanvasObject {
	return r.buildFreeFormLine(msg, r.config.FreeFormColorParsed())
}

func (r *LogLineRenderer) buildRenderedLine(msg string, borderColor, badgeColor color.Color, badgeText string) fyne.CanvasObject {
	bar   := r.buildLeftBar(borderColor)
	badge := r.buildBadge(badgeText, badgeColor)
	label := widget.NewLabelWithStyle(msg, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	return container.NewHBox(bar, badge, label)
}

func (r *LogLineRenderer) buildFreeFormLine(msg string, borderColor color.Color) fyne.CanvasObject {
	bar   := r.buildLeftBar(borderColor)
	label := widget.NewLabelWithStyle(msg, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	return container.NewHBox(bar, label)
}

func (r *LogLineRenderer) buildLeftBar(c color.Color) fyne.CanvasObject {
	rect := canvas.NewRectangle(c)
	rect.SetMinSize(fyne.NewSize(4, 20))
	return rect
}

func (r *LogLineRenderer) buildBadge(text string, c color.Color) fyne.CanvasObject {
	bg    := canvas.NewRectangle(c)
	label := canvas.NewText(text, color.White)
	label.TextSize  = 10
	label.TextStyle = fyne.TextStyle{Bold: true}
	bg.SetMinSize(fyne.NewSize(float32(len(text))*7+8, 18))
	return container.NewStack(bg, container.NewCenter(label))
}
