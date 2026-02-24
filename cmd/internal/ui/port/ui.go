// ===============================
package port

type UIObject interface {
	Native() any
	Title() string
	Description() string
	ToggleItem(*LegendItem)
}

type TabItem interface {
	Native() any
}

type LegendItem struct {
	Key   string // stable id (e.g. "cpu")
	Label string // display label (e.g. "CPU Metrics")
	Color string // e.g. "dodgerblue" or "#1e90ff"
	Index int
}

// Layout builds UI containers (VBox, Grid, Tabs, etc.) without exposing toolkit details.
type Layout interface {
	Card(title, subtitle string, content UIObject) UIObject
	VBox(children ...UIObject) UIObject
	Max(children ...UIObject) UIObject
	GridCols(cols int, children ...UIObject) UIObject
	Padded(child UIObject) UIObject

	Tabs(items ...TabItem) UIObject
	Tab(title string, content UIObject) TabItem

	Label(text string) UIObject
	Title(text string) UIObject
	Separator() UIObject

	VScroll(content UIObject) UIObject

	// NEW primitives
	HBox(children ...UIObject) UIObject
	Spacer() UIObject

	// NEW: legend for multi-variable diagrams.
	// If onClick is nil => non-interactive.
	DiagramLegend(items []LegendItem, onClick func(*LegendItem)) UIObject

	SetLegendAction(legend UIObject, onClick func(*LegendItem))

	Border(top, bottom, left, right, center UIObject) UIObject
}
