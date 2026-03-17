package fynediagrams

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// lastUpdatedLabel is a thin UIObject wrapper around a Fyne label that tracks
// when data was last pushed to a diagram.
type lastUpdatedLabel struct {
	label *widget.Label
}

func newLastUpdatedLabel() *lastUpdatedLabel {
	lbl := widget.NewLabel("")
	lbl.Alignment = fyne.TextAlignCenter
	lbl.TextStyle = fyne.TextStyle{Italic: true}
	return &lastUpdatedLabel{label: lbl}
}

// update sets the label text on the UI thread.
func (l *lastUpdatedLabel) update(t time.Time) {
	UI(func() {
		l.label.SetText(formatLastUpdated(t))
	})
}

// uport.UIObject implementation
func (l *lastUpdatedLabel) Native() any                    { return l.label }
func (l *lastUpdatedLabel) Title() string                  { return "" }
func (l *lastUpdatedLabel) Description() string            { return "" }
func (l *lastUpdatedLabel) ToggleItem(*uport.LegendItem)   {}
func (l *lastUpdatedLabel) Refresh()                       { l.label.Refresh() }

// formatLastUpdated returns the display string for the given timestamp.
// Same day  → "Last updated today at 3:04pm"
// Other day → "Last updated: 17/3/2026 3:04pm"
func formatLastUpdated(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return "Last updated today at " + t.Format("3:04pm")
	}
	return "Last updated: " + t.Format("2/1/2006 3:04pm")
}
