package ui

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gclkaze/evamon/cmd/internal/models"
	window "github.com/gclkaze/evamon/cmd/internal/ui/chart"
	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/magiconair/properties"
)

type ProjectUIHolder struct {
	vp *viewproject.ViewProject
	//can host multiple tabs and one window
	//or multiple windows
	windows       []port.ExecutionWindow
	renderer      port.Renderer
	drawerFactory draw.Factory
	//ToolbarFactory: toolbarFactory,
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

	var windowStyle *models.WindowStyle
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

func (inst *ProjectUIHolder) setupAndBuildMultiTabbedDiagramWindow(diagrams []models.Diagram, width *float32, height *float32) ([]port.ExecutionWindow, error) {

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

func (inst *ProjectUIHolder) setupAndBuildDiagramsMultiTabed(diagram models.Diagram, width *float32, height *float32) (port.ExecutionWindow, error) {
	theSetupItems := diagram.Setup
	w, err := inst.renderer.NewExecutionWindow("")
	if err != nil {
		return nil, err
	}
	w = inst.styleWindow(w, width, height, nil, nil)
	ref := port.NewDiagramUIRefs(diagram.ID, inst.renderer)
	for j := range theSetupItems {
		setup := theSetupItems[j]
		title := setup.Title
		tabID := strconv.Itoa(j)
		w.UpsertTab(tabID, title)

		content := window.BuildDiagramContent(inst, inst.drawerFactory, w, setup, inst.MinChartWidth, inst.MinChartHeight, inst.defaultMaxPoints, diagram.Type, ref, &diagram)
		w.AssignTab(tabID, content)
		w.Show()
	}

	w.CommitTabs()
	return w, nil
}

func (inst *ProjectUIHolder) GetProjectID() string {
	return inst.vp.ID
}

func (inst *ProjectUIHolder) setupAndBuildDiagrams(diagram models.Diagram, windowStyle *models.WindowStyle, width *float32, height *float32) ([]port.ExecutionWindow, error) {
	theSetupItems := diagram.Setup
	var ws []port.ExecutionWindow
	for j := range theSetupItems {
		setup := theSetupItems[j]

		w, err := inst.renderer.NewExecutionWindow(setup.Title)
		if err != nil {
			return nil, err
		}

		w = inst.styleWindow(w, width, height, &setup, windowStyle)
		w = window.BuildDiagram(inst, inst.drawerFactory, w, setup, inst.MinChartWidth, inst.MinChartHeight, inst.defaultMaxPoints, inst.renderer, &diagram)

		w.Show()

		ws = append(ws, w)
	}
	return ws, nil
}

func (inst *ProjectUIHolder) GetRenderer() port.Renderer {
	return inst.renderer
}

func (inst *ProjectUIHolder) RegisterVariableDrawerUnsubscriber(variableName string, drawer draw.DiagramWidget) {
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

func (inst *ProjectUIHolder) styleWindow(window port.ExecutionWindow, w *float32, h *float32, setup *models.SetupItem, windowStyle *models.WindowStyle) port.ExecutionWindow {

	if setup.WindowStyle != nil {
		w = setup.WindowStyle.Width
		h = setup.WindowStyle.Height
	} else {
		if windowStyle != nil {
			w = windowStyle.Width
			h = windowStyle.Height
		}
	}

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
