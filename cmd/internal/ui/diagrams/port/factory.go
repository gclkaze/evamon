package port

import (
	"image/color"
)

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

	//	MultiColor []color.Color

	Variables []VariableStyle
}
type VariableStyle struct {
	VariableName string
	VarColor     color.Color
	VarColorText string
}

// LineChartOptions controls rendering and scaling.
type LineChartOptions struct {
	MaxPoints int

	// Visual toggles
	ShowAxes    bool
	ShowGrid    bool
	ShowMarkers bool

	// Scaling
	YPadRatio float64 // e.g. 0.10 adds 10% padding above/below min/max

	// Plot padding inside widget
	PadL float32
	PadR float32
	PadT float32
	PadB float32

	// Colors
	Background color.Color
	Axis       color.Color
	Grid       color.Color
	//Line       color.Color
	Marker color.Color

	// Stroke widths
	AxisStroke   float32
	GridStroke   float32
	LineStroke   float32
	MarkerRadius float32

	// Grid density
	GridX int // vertical grid lines
	GridY int // horizontal grid lines

	Variables []VariableStyle
}

func DefaultLineChartOptions() LineChartOptions {
	return LineChartOptions{
		MaxPoints: 120,

		ShowAxes:    true,
		ShowGrid:    true,
		ShowMarkers: true,

		YPadRatio: 0.10,

		PadL: 44,
		PadR: 12,
		PadT: 12,
		PadB: 28,

		Background: color.NRGBA{R: 15, G: 15, B: 15, A: 255},
		Axis:       color.NRGBA{R: 190, G: 190, B: 190, A: 255},
		Grid:       color.NRGBA{R: 70, G: 70, B: 70, A: 255},
		//Line:       color.NRGBA{R: 80, G: 130, B: 255, A: 255},
		Marker: color.NRGBA{R: 220, G: 220, B: 220, A: 255},

		AxisStroke:   1,
		GridStroke:   1,
		LineStroke:   2,
		MarkerRadius: 3,

		GridX: 6,
		GridY: 4,
	}
}

type Factory interface {
	NewBoolFill(opts BoolFillOptions, title string, description string, width, height float32, varname string) DiagramWidget
	NewBarChart(opts BarChartOptions, title string, description string, width, height float32) DiagramWidget
	NewLineChart(opts LineChartOptions, title string, description string, initialWidth, initialHeight, width, height float32) DiagramWidget
}
