package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"sync"

	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/magiconair/properties"
	"golang.org/x/image/colornames"
)

type ProjectUIHolder struct {
	vp *viewproject.ViewProject
	//can host multiple tabs and one window
	//or multiple windows
	windows       []port.ExecutionWindow
	renderer      port.Renderer
	drawerFactory draw.Factory

	isMultitab bool
	props      *properties.Properties

	defaultWidth     float32
	defaultHeight    float32
	defaultResizable bool
	defaultMaxPoints int

	vc *VariableContainer

	mu             sync.RWMutex
	unsubscribe    map[string]map[draw.DiagramWidget]func()
	MinChartWidth  float32
	MinChartHeight float32
}

func NewProjectUIHolder(vp *viewproject.ViewProject, renderer port.Renderer, drawerFactory draw.Factory, props *properties.Properties, vc *VariableContainer) *ProjectUIHolder {
	//one window holder per view project
	return &ProjectUIHolder{vp: vp, renderer: renderer, props: props, drawerFactory: drawerFactory, vc: vc, unsubscribe: make(map[string]map[draw.DiagramWidget]func()), MinChartWidth: 300, MinChartHeight: 300}
}

func (inst *ProjectUIHolder) SetOnClosed(close func()) {
	for i := range inst.windows {
		inst.windows[i].SetOnClosed(close)
	}
}

func (inst *ProjectUIHolder) Create() error {
	if inst.vp == nil {
		return fmt.Errorf("no project has been set")
	}

	inst.loadDefaults()

	inst.isMultitab = inst.vp.View.MultiTab

	var windowStyle *viewproject.WindowStyle
	var height *float32
	var width *float32

	windowStyle = inst.vp.View.WindowStyle
	if windowStyle != nil {
		height = windowStyle.Height
		width = windowStyle.Width
	}

	if inst.isMultitab {
		//for now, one window, multiple tabs
		diagrams := inst.vp.View.Diagrams
		ws, err := inst.setupAndBuildMultiTabbedDiagramWindow(diagrams, width, height)
		if err != nil {
			return err
		}
		inst.windows = ws
	} else {
		diagrams := inst.vp.View.Diagrams
		for i := range diagrams {
			ws, err := inst.setupAndBuildDiagrams(diagrams[i], windowStyle, width, height)
			if err != nil {
				return err
			}

			inst.windows = ws
		}
	}

	return nil
}

func (inst *ProjectUIHolder) Run() {
	inst.renderer.Run()
}

func (inst *ProjectUIHolder) setupAndBuildMultiTabbedDiagramWindow(diagrams []viewproject.Diagram, width *float32, height *float32) ([]port.ExecutionWindow, error) {

	var ws []port.ExecutionWindow
	for i := range diagrams {
		w, err := inst.setupAndBuildDiagramsMultiTabed(diagrams[i], width, height)
		if err != nil {
			return nil, err
		}

		ws = append(ws, w)
	}
	fmt.Printf("%d", len(ws))

	return ws, nil
}

func (inst *ProjectUIHolder) setupAndBuildDiagramsMultiTabed(diagram viewproject.Diagram, width *float32, height *float32) (port.ExecutionWindow, error) {
	theSetupItems := diagram.Setup
	w, err := inst.renderer.NewExecutionWindow("")
	if err != nil {
		return nil, err
	}
	w = inst.styleWindow(w, width, height)

	for j := range theSetupItems {
		setup := theSetupItems[j]
		title := setup.Title
		tabID := strconv.Itoa(j)
		w.UpsertTab(tabID, title)

		content := inst.buildDiagramContent(setup)
		w.AssignTab(tabID, content)
		w.Show()
	}

	w.CommitTabs()
	return w, nil
}

func (inst *ProjectUIHolder) setupAndBuildDiagrams(diagram viewproject.Diagram, windowStyle *viewproject.WindowStyle, width *float32, height *float32) ([]port.ExecutionWindow, error) {
	theSetupItems := diagram.Setup
	var ws []port.ExecutionWindow
	for j := range theSetupItems {
		setup := theSetupItems[j]

		w, err := inst.renderer.NewExecutionWindow(setup.Title)
		if err != nil {
			return nil, err
		}

		if setup.WindowStyle != nil {
			width = setup.WindowStyle.Width
			height = setup.WindowStyle.Height
		} else {
			if windowStyle != nil {
				width = windowStyle.Width
				height = windowStyle.Height
			}
		}

		w = inst.styleWindow(w, width, height)
		w = inst.buildDiagram(w, setup)
		w.Show()

		ws = append(ws, w)
	}
	return ws, nil
}
func (inst *ProjectUIHolder) buildDiagram(w port.ExecutionWindow, setup viewproject.SetupItem) port.ExecutionWindow {
	switch setup.VariableType {
	case viewproject.ValueTypeBoolean:

		boolFill := inst.buildBooleanDiagram(setup)

		w.SetContent(boolFill)
		w.SetResizable(true)

	case viewproject.ValueTypeInteger:

		barChart := inst.buildBarchart(setup)

		w.SetContent(barChart)
		w.SetResizable(true)
	}
	return w
}

func (inst *ProjectUIHolder) buildBooleanDiagram(setup viewproject.SetupItem) draw.DiagramWidget {
	var falseColor color.RGBA
	var trueColor color.RGBA
	if setup.DiagramStyle != nil {
		if bs, ok := setup.DiagramStyle.(viewproject.BooleanStyle); ok {
			trueColor = colornames.Map[bs.True]
			falseColor = colornames.Map[bs.False]
		}
	} else {
		falseColor = colornames.Map["red"]
		trueColor = colornames.Map["green"]

	}

	boolFill := inst.drawerFactory.NewBoolFill(draw.BoolFillOptions{
		TrueColor:  trueColor,
		FalseColor: falseColor,
	}, setup.Title, setup.Description, inst.MinChartWidth, inst.MinChartHeight)

	inst.registerVariableDrawerUnsubscriber(setup.Variable, boolFill)
	return boolFill
}

func (inst *ProjectUIHolder) buildBarchart(setup viewproject.SetupItem) draw.DiagramWidget {
	var axisColor color.RGBA
	var backgroundColor color.RGBA
	if setup.DiagramStyle != nil {
		if bs, ok := setup.DiagramStyle.(viewproject.BarStyle); ok {
			axisColor = colornames.Map[bs.Axis]
			backgroundColor = colornames.Map[bs.Background]
		}
	} else {
		axisColor = colornames.Map["red"]
		backgroundColor = colornames.Map["green"]
	}

	barchart := inst.drawerFactory.NewBarChart(draw.BarChartOptions{
		MaxPoints:  inst.defaultMaxPoints,
		Width:      0,
		Height:     0,
		Axis:       axisColor,
		Background: backgroundColor,
	}, setup.Title, setup.Description, inst.MinChartWidth, inst.MinChartHeight)

	inst.registerVariableDrawerUnsubscriber(setup.Variable, barchart)
	return barchart
}

func (inst *ProjectUIHolder) buildDiagramContent(setup viewproject.SetupItem) draw.DiagramWidget {
	switch setup.VariableType {
	case viewproject.ValueTypeBoolean:
		boolFill := inst.buildBooleanDiagram(setup)
		return boolFill
	case viewproject.ValueTypeInteger:
		barChart := inst.buildBarchart(setup)
		return barChart

	}
	return nil
}

func (inst *ProjectUIHolder) registerVariableDrawerUnsubscriber(variableName string /* drawer draw.Drawer*/, drawer draw.DiagramWidget) {
	rem := inst.vc.Register(variableName, drawer)
	inst.mu.Lock()
	set := inst.unsubscribe[variableName]
	if set == nil {
		set = make(map[draw.DiagramWidget]func())
		inst.unsubscribe[variableName] = set
	}

	set[drawer] = func() {}
	inst.unsubscribe[variableName][drawer] = rem
	inst.mu.Unlock()
}

func (inst *ProjectUIHolder) styleWindow(window port.ExecutionWindow, w *float32, h *float32) port.ExecutionWindow {

	if w != nil && h != nil {
		window.Resize(float32(*w), float32(*h))
	} else {
		window.Resize(inst.defaultWidth, inst.defaultHeight)
	}

	window.SetResizable(inst.defaultResizable)
	return window
}

func (inst *ProjectUIHolder) Update() error {
	return nil
}

func (inst *ProjectUIHolder) loadDefaults() {
	if inst.props == nil {
		return
	}

	inst.defaultWidth = inst.props.GetFloat32("default_window_width", 400)
	inst.defaultHeight = inst.props.GetFloat32("default_window_height", 400)
	inst.defaultResizable = inst.props.GetBool("default_window_resizable", false)

	inst.defaultMaxPoints = inst.props.GetInt("default_barchart_maxpoints", 1000)
}
