package port

import "image/color"

// Kind identifies which drawer to create.
type Kind string

const (
	KindBoolFill Kind = "bool_fill"
	KindBarChart Kind = "barchart"
)

type BoolFillOptions struct {
	TrueColor  color.Color
	FalseColor color.Color
}

type BarChartOptions struct {
	MaxPoints  int
	Width      float32
	Height     float32
	Axis       color.Color
	Background color.Color
}

type Factory interface {
	NewBoolFill(opts BoolFillOptions) DiagramWidget
	NewBarChart(opts BarChartOptions) DiagramWidget
}
