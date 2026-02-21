package port

// UIObject is an opaque UI node.
// Concrete renderers (fyne, web, etc.) wrap their native objects behind this.
type UIObject interface {
	Native() any
	Title() string
	Description() string
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
}

type TabItem interface {
	Native() any
}
