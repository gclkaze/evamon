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

	var theVariables []draw.VariableStyle
	if setup.MultiVariableSetup != nil {
		vars := make([]draw.VariableStyle, 0)
		for i := range setup.MultiVariableSetup {
			current := setup.MultiVariableSetup[i]
			theBar, ok := current.DiagramStyle.(viewproject.BarStyle)
			if !ok {
				continue
			}
			col := colornames.Map[theBar.Bar]
			vars = append(vars, draw.VariableStyle{VariableName: current.Variable, VarColor: col})
		}
		theVariables = vars
	} else {
		if bs, ok := setup.DiagramStyle.(viewproject.BarStyle); ok {
			barColor := colornames.Map[bs.Axis]
			theVariables = []draw.VariableStyle{{VariableName: setup.Variable, VarColor: barColor}}
		}
	}
	w := drawerFactory.NewBarChart(draw.BarChartOptions{
		Width:      0,
		Height:     0,
		MaxPoints:  maxPoints,
		Axis:       axisColor,
		Background: backgroundColor,
		Variables:  theVariables,
	}, setup.Title, setup.Description, width, height)
	return w, nil
}

func CreateLinechartDrawer(drawerFactory draw.Factory, setup *viewproject.SetupItem, index int, width, height float32, maxPoints int) (draw.DiagramWidget, error) {
	opts := draw.DefaultLineChartOptions()

	if setup.MultiVariableSetup != nil {
		vars := make([]draw.VariableStyle, 0)
		for i := range setup.MultiVariableSetup {
			current := setup.MultiVariableSetup[i]
			theBar, ok := current.DiagramStyle.(viewproject.LineStyle)
			if !ok {
				continue
			}
			col := colornames.Map[theBar.Line]
			vars = append(vars, draw.VariableStyle{VariableName: current.Variable, VarColor: col})
		}
		opts.Variables = vars
	} else {
		if bs, ok := setup.DiagramStyle.(viewproject.LineStyle); ok {
			lineColor := colornames.Map[bs.Line]
			backgroundColor := colornames.Map[bs.Background]
			opts.Background = backgroundColor

			opts.Variables = []draw.VariableStyle{{VariableName: setup.Variable, VarColor: lineColor}}
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
