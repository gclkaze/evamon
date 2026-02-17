package port

// UIObject is an opaque UI node.
// Concrete renderers (fyne, web, etc.) wrap their native objects behind this.
type UIObject interface {
	Native() any
}

// Layout builds UI containers (VBox, Grid, Tabs, etc.) without exposing toolkit details.
type Layout interface {
	VBox(children ...UIObject) UIObject
	Max(children ...UIObject) UIObject
	GridCols(cols int, children ...UIObject) UIObject
	Padded(child UIObject) UIObject

	Tabs(items ...TabItem) UIObject
	Tab(title string, content UIObject) TabItem
}

type TabItem interface {
	Native() any
}
