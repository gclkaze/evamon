package window

import (
	"fmt"
	"image/color"
	"strings"

	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"golang.org/x/image/colornames"
)

func wrapWithDiagramPanel(toolbar, content port.UIObject, renderer port.Renderer) port.UIObject {
	if toolbar == nil {
		return content
	}
	if content == nil {
		return toolbar
	}

	return renderer.Layout().Border(
		toolbar,
		nil,
		nil,
		nil,
		content,
	)
}

func BuildDiagramContent(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, w port.ExecutionWindow, setup viewproject.SetupItem, width, height float32, maxPoints int, t viewproject.DiagramType) draw.DiagramWidget {
	/*	tf := &ui.DefaultDiagramToolbarFactory{
		Layout:   holder.GetRenderer().Layout(),
		Controls: holder.GetRenderer().Controls(),
		Actions:  holder.GetRenderer().Actions(),
	}*/

	switch t {
	case viewproject.DiagramTypeBoolean:
		boolFill := buildBooleanDiagram(holder, drawerFactory, setup, width, height)
		//toolbar := tf.Build(jobID, boolFill)
		//panel := wrapWithDiagramPanel(toolbar, content)
		return boolFill
	case viewproject.DiagramTypeBar:
		barChart := buildBarchart(holder, drawerFactory, setup, width, height, maxPoints)
		return barChart
	case viewproject.DiagramTypeLine:
		lineChart := buildLinechart(holder, drawerFactory, setup, width, height, maxPoints)
		return lineChart
	}
	return nil
}

func BuildDiagram(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, w port.ExecutionWindow, setup viewproject.SetupItem, width, height float32, maxPoints int,
	renderer port.Renderer, diagram *viewproject.Diagram) port.ExecutionWindow {

	vars := CollectVariables(diagram)
	theItems := VariableStylesToLegendItems(vars)

	drawer := BuildDiagramContent(holder, drawerFactory, w, setup, width, height, maxPoints, diagram.Type)

	l := renderer.Layout()
	legendObj := l.DiagramLegend(theItems, func(it *port.LegendItem) {
		if len(theItems) == 1 {
			return
		}
		drawer.ToggleItem(it)
	})

	c := l.Border(legendObj, nil, nil, nil, drawer)
	w.SetContent(c)
	w.SetResizable(true)
	return w
}

func VariableStylesToLegendItems(vars []draw.VariableStyle) []port.LegendItem {
	out := make([]port.LegendItem, 0, len(vars))

	for i, v := range vars {
		out = append(out, port.LegendItem{
			Key:   v.VariableName, // stable id
			Label: v.VariableName, // you can change if you later add display name
			Color: v.VarColorText, // convert color.Color → string
			Index: i,
		})
	}

	return out
}

func CollectVariables(d *viewproject.Diagram) []draw.VariableStyle {
	var out []draw.VariableStyle

	for si := range d.Setup {
		s := &d.Setup[si]

		// ----------------------------
		// Multi-variable
		// ----------------------------
		if len(s.MultiVariableSetup) > 0 {
			for ci := range s.MultiVariableSetup {
				ms := &s.MultiVariableSetup[ci]

				var col color.Color
				txt := ""
				switch d.Type {
				case viewproject.DiagramTypeBar:
					if bs, ok := ms.DiagramStyle.(viewproject.BarStyle); ok {
						txt = bs.Bar
						col = lookupColor(bs.Bar)
					}

				case viewproject.DiagramTypeLine:
					if ls, ok := ms.DiagramStyle.(viewproject.LineStyle); ok {
						txt = ls.Line
						col = lookupColor(ls.Line)
					}
				}

				out = append(out, draw.VariableStyle{
					VariableName: ms.Variable,
					VarColor:     col,
					VarColorText: txt,
				})
			}
			continue
		}

		// ----------------------------
		// Single-variable
		// ----------------------------
		var col color.Color
		txt := ""
		switch d.Type {
		case viewproject.DiagramTypeBar:
			if bs, ok := s.DiagramStyle.(viewproject.BarStyle); ok {
				txt = bs.Axis
				col = lookupColor(bs.Axis)
			}

		case viewproject.DiagramTypeLine:
			if ls, ok := s.DiagramStyle.(viewproject.LineStyle); ok {
				txt = ls.Line
				col = lookupColor(ls.Line)
			}

		case viewproject.DiagramTypeBoolean:
			if bs, ok := s.DiagramStyle.(viewproject.BooleanStyle); ok {
				txt = bs.True
				col = lookupColor(bs.True)
			}
		}

		out = append(out, draw.VariableStyle{
			VariableName: s.Variable,
			VarColor:     col,
			VarColorText: txt,
		})
	}

	return out
}
func lookupColor(s string) color.Color {
	s = strings.ToLower(strings.TrimSpace(s))

	if c, ok := colornames.Map[s]; ok {
		return c
	}

	// fallback (neutral gray if unknown)
	return colornames.Gray
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
	}, setup.Title, setup.Description, width, height, setup.Variable)
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
			vars = append(vars, draw.VariableStyle{VariableName: current.Variable, VarColor: col, VarColorText: theBar.Line})
		}
		opts.Variables = vars
	} else {
		if bs, ok := setup.DiagramStyle.(viewproject.LineStyle); ok {
			lineColor := colornames.Map[bs.Line]
			backgroundColor := colornames.Map[bs.Background]
			opts.Background = backgroundColor

			opts.Variables = []draw.VariableStyle{{VariableName: setup.Variable, VarColor: lineColor, VarColorText: bs.Line}}
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

func buildLinechart(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup viewproject.SetupItem, width, height float32, maxPoints int) draw.DiagramWidget {
	linechart, err := CreateLinechartDrawer(drawerFactory, &setup, -1, width, height, maxPoints)
	if err != nil {
		return nil
	}

	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, linechart)
	return linechart
}

func buildBarchart(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup viewproject.SetupItem, width, height float32, maxPoints int) draw.DiagramWidget {
	barchart, err := CreateBarchartDrawer(drawerFactory, &setup, -1, width, height, maxPoints)
	if err != nil {
		return nil
	}

	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, barchart)
	return barchart
}
