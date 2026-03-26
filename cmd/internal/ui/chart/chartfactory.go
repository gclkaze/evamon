package window

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/gclkaze/evamon/cmd/internal/models"
	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	ui "github.com/gclkaze/evamon/cmd/internal/ui/factory"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	"golang.org/x/image/colornames"
)

func BuildDiagramContent(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, w port.ExecutionWindow, setup models.SetupItem, width, height float32, maxPoints int, t models.DiagramType, ref *port.DiagramUIRefs, d port.IDiagram) draw.DiagramWidget {
	switch t {
	case models.DiagramTypeBoolean:
		boolFill := buildBooleanDiagram(holder, drawerFactory, setup, width, height, d)
		ref.Chart = boolFill
		return boolFill
	case models.DiagramTypeBar:
		barChart := buildBarchart(holder, drawerFactory, setup, width, height, maxPoints, d)
		ref.Chart = barChart

		return barChart
	case models.DiagramTypeLine:
		lineChart := buildLinechart(holder, drawerFactory, setup, width, height, maxPoints, d)
		ref.Chart = lineChart
		return lineChart
	}
	return nil
}

func BuildDiagram(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, w port.ExecutionWindow, setup models.SetupItem, width, height float32, maxPoints int,
	renderer port.Renderer, diagram *models.Diagram) port.ExecutionWindow {

	vars := CollectVariables(diagram)
	theItems := VariableStylesToLegendItems(vars)

	ref := port.NewDiagramUIRefs(diagram.ID, holder.GetRenderer())

	drawer := BuildDiagramContent(holder, drawerFactory, w, setup, width, height, maxPoints, diagram.Type, ref, diagram)

	l := renderer.Layout()
	legendObj := l.DiagramLegend(theItems, func(it *port.LegendItem) {
		if len(theItems) == 1 {
			return
		}
		drawer.ToggleItem(it)
	})

	tf := ui.NewDiagramToolbarFactor(holder.GetRenderer())

	renderer.ChartRegistry().Register(ref)
	toolbar := tf.Build("", diagram)
	ref.RegisterToolbar(toolbar)
	ref.Toolbar = toolbar
	ref.ID = diagram.ID
	ref.Legend = legendObj
	ref.RegisterLegend(legendObj)
	ref.RegisterMainChart(drawer)

	//	panel := wrapWithDiagramPanel(toolbar, content)
	fmt.Println("Registering " + ref.ID)
	fmt.Printf("Size of Registry is %d \n", renderer.ChartRegistry().Size())

	fmt.Printf("Size of Registry is %d \n", renderer.ChartRegistry().Size())

	var bottom port.UIObject
	if toolbar != nil {
		bottom = l.VBox(toolbar, drawer.LastUpdatedLabel())
	} else {
		bottom = l.VBox(drawer.LastUpdatedLabel())
	}
	c := l.Border(legendObj, bottom, nil, nil, drawer)
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

func CollectVariables(d *models.Diagram) []draw.VariableStyle {
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
				case models.DiagramTypeBar:
					if bs, ok := ms.DiagramStyle.(models.BarStyle); ok {
						txt = bs.Bar
						col = lookupColor(bs.Bar)
					}

				case models.DiagramTypeLine:
					if ls, ok := ms.DiagramStyle.(models.LineStyle); ok {
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
		case models.DiagramTypeBar:
			if bs, ok := s.DiagramStyle.(models.BarStyle); ok {
				txt = bs.Axis
				col = lookupColor(bs.Axis)
			}

		case models.DiagramTypeLine:
			if ls, ok := s.DiagramStyle.(models.LineStyle); ok {
				txt = ls.Line
				col = lookupColor(ls.Line)
			}

		case models.DiagramTypeBoolean:
			if bs, ok := s.DiagramStyle.(models.BooleanStyle); ok {
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

func buildBooleanDiagram(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup models.SetupItem, width, height float32, owner port.IDiagram) draw.DiagramWidget {
	boolFill, err := CreateBoolDrawer(drawerFactory, &setup, -1, width, height, owner)
	if err != nil {
		return nil
	}
	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, boolFill)
	return boolFill
}

func CreateBoolDrawer(drawerFactory draw.Factory, setup *models.SetupItem, index int, width, height float32, owner port.IDiagram) (draw.DiagramWidget, error) {
	// Extract style if present
	var trueColor, falseColor color.RGBA
	if setup.DiagramStyle != nil {
		st, ok := setup.DiagramStyle.(models.BooleanStyle)
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
	}, setup.Title, setup.Description, width, height, setup.Variable, owner)
	return w, nil
}

func CreateBarchartDrawer(drawerFactory draw.Factory, setup *models.SetupItem, index int, width, height float32, maxPoints int, d port.IDiagram) (draw.DiagramWidget, error) {
	var axisColor, backgroundColor color.RGBA

	if setup.DiagramStyle != nil {
		if bs, ok := setup.DiagramStyle.(models.BarStyle); ok {
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
			theBar, ok := current.DiagramStyle.(models.BarStyle)
			if !ok {
				continue
			}
			col := colornames.Map[theBar.Bar]
			vars = append(vars, draw.VariableStyle{VariableName: current.Variable, VarColor: col})
		}
		theVariables = vars
	} else {
		if bs, ok := setup.DiagramStyle.(models.BarStyle); ok {
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
	}, setup.Title, setup.Description, width, height, d)
	return w, nil
}

func CreateLinechartDrawer(drawerFactory draw.Factory, setup *models.SetupItem, index int, width, height float32, maxPoints int, d port.IDiagram) (draw.DiagramWidget, error) {
	opts := draw.DefaultLineChartOptions()

	if setup.MultiVariableSetup != nil {
		vars := make([]draw.VariableStyle, 0)
		for i := range setup.MultiVariableSetup {
			current := setup.MultiVariableSetup[i]
			theBar, ok := current.DiagramStyle.(models.LineStyle)
			if !ok {
				continue
			}
			col := colornames.Map[theBar.Line]
			vars = append(vars, draw.VariableStyle{VariableName: current.Variable, VarColor: col, VarColorText: theBar.Line})
		}
		opts.Variables = vars
	} else {
		if bs, ok := setup.DiagramStyle.(models.LineStyle); ok {
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
		d,
	)

	return w, nil
}

func buildLinechart(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup models.SetupItem, width, height float32, maxPoints int, d port.IDiagram) draw.DiagramWidget {
	linechart, err := CreateLinechartDrawer(drawerFactory, &setup, -1, width, height, maxPoints, d)
	if err != nil {
		return nil
	}

	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, linechart)
	return linechart
}

func buildBarchart(holder draw.VariableDrawerOwner, drawerFactory draw.Factory, setup models.SetupItem, width, height float32, maxPoints int, d port.IDiagram) draw.DiagramWidget {
	barchart, err := CreateBarchartDrawer(drawerFactory, &setup, -1, width, height, maxPoints, d)
	if err != nil {
		return nil
	}

	holder.RegisterVariableDrawerUnsubscriber(setup.Variable, barchart)
	return barchart
}
