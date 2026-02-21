package window

import (
	"fmt"
	"image/color"

	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"golang.org/x/image/colornames"
)

func BuildDiagramContent(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, w port.ExecutionWindow, setup viewproject.SetupItem, width, height float32, maxPoints int) draw.DiagramWidget {
	switch setup.VariableType {
	case viewproject.ValueTypeBoolean:
		boolFill := buildBooleanDiagram(holder, drawerFactory, setup, width, height)
		return boolFill
	case viewproject.ValueTypeInteger:
		barChart := buildBarchart(holder, drawerFactory, setup, width, height, maxPoints)
		return barChart
	}
	return nil
}

func BuildDiagram(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, w port.ExecutionWindow, setup viewproject.SetupItem, width, height float32, maxPoints int) port.ExecutionWindow {
	drawer := BuildDiagramContent(holder, drawerFactory, w, setup, width, height, maxPoints)
	w.SetContent(drawer)
	w.SetResizable(true)
	return w
}

func buildBooleanDiagram(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup viewproject.SetupItem, width, height float32) draw.DiagramWidget {
	boolFill, err := CreateBoolDrawer(drawerFactory, &setup, -1, width, height)
	if err != nil {
		return nil
	}
	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, boolFill)
	return boolFill
}

func CreateBoolDrawer(drawerFactory draw.Factory, setup *viewproject.SetupItem, index int, width, height float32) (draw.DiagramWidget, error) {
	// Extract style if present
	var trueColor, falseColor color.RGBA
	if setup.DiagramStyle != nil {
		st, ok := setup.DiagramStyle.(viewproject.BooleanStyle)
		if !ok {
			if index == -1 {
				return nil, fmt.Errorf("diagramStyle is not BooleanStyle")
			}
			return nil, fmt.Errorf("setup[%d]: diagramStyle is not BooleanStyle", index)
		}
		trueColor = colornames.Map[st.True]
		falseColor = colornames.Map[st.False]
	} else {
		falseColor = colornames.Map["red"]
		trueColor = colornames.Map["green"]
	}

	// Build widget; since we render title outside, avoid duplicating it inside the widget
	w := drawerFactory.NewBoolFill(draw.BoolFillOptions{
		TrueColor:  trueColor,
		FalseColor: falseColor,
	}, setup.Title, setup.Description, width, height)
	return w, nil
}

func CreateBarchartDrawer(drawerFactory draw.Factory, setup *viewproject.SetupItem, index int, width, height float32, maxPoints int) (draw.DiagramWidget, error) {
	var axisColor, backgroundColor color.RGBA

	if setup.DiagramStyle != nil {
		if bs, ok := setup.DiagramStyle.(viewproject.BarStyle); ok {
			axisColor = colornames.Map[bs.Axis]
			backgroundColor = colornames.Map[bs.Background]
			if !ok {
				if index == -1 {
					return nil, fmt.Errorf("diagramStyle is not BarStyle")
				}

				return nil, fmt.Errorf("setup[%d]: diagramStyle is not BarStyle", index)
			}
		}
	} else {
		axisColor = colornames.Map["red"]
		backgroundColor = colornames.Map["black"]
	}

	w := drawerFactory.NewBarChart(draw.BarChartOptions{
		// If you have per-setup windowStyle, you can map it here too.
		// Width/Height are typically 0 so layout controls the size.
		Width:      0,
		Height:     0,
		MaxPoints:  maxPoints, // or set from somewhere else (global default)
		Axis:       axisColor,
		Background: backgroundColor,
	}, setup.Title, setup.Description, width, height)
	return w, nil
}

func CreateLinechartDrawer(drawerFactory draw.Factory, setup *viewproject.SetupItem, index int, width, height float32, maxPoints int) (draw.DiagramWidget, error) {
	opts := draw.DefaultLineChartOptions()

	if setup.DiagramStyle != nil {
		if bs, ok := setup.DiagramStyle.(viewproject.LineStyle); ok {
			lineColor := colornames.Map[bs.Line]
			backgroundColor := colornames.Map[bs.Background]

			opts.Line = lineColor
			opts.Background = backgroundColor
		}
	}

	w := drawerFactory.NewLineChart(
		opts,
		setup.Title,
		setup.Description,
		800, 260, // initial raster (so it doesn’t start tiny/blurry)
		width, height, // widget hint; layout will expand it
	)

	return w, nil
}

func buildBarchart(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup viewproject.SetupItem, width, height float32, maxPoints int) draw.DiagramWidget {
	barchart, err := CreateBarchartDrawer(drawerFactory, &setup, -1, width, height, maxPoints)
	if err != nil {
		return nil
	}

	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, barchart)
	return barchart
}
