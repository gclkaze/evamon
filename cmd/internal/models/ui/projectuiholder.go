package ui

import (
	"fmt"
	"image/color"
	"strconv"

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
}

func NewProjectUIHolder(vp *viewproject.ViewProject, renderer port.Renderer, drawerFactory draw.Factory, props *properties.Properties) *ProjectUIHolder {
	//one window holder per view project
	return &ProjectUIHolder{vp: vp, renderer: renderer, props: props, drawerFactory: drawerFactory}
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

	if inst.isMultitab {
		//for now, one window, multiple tabs
		windowStyle = inst.vp.View.WindowStyle
		if windowStyle != nil {
			height = windowStyle.Height
			width = windowStyle.Width
		}

		diagrams := inst.vp.View.Diagrams
		ws, err := inst.setupAndBuildMultiTabbedDiagramWindow(diagrams, width, height)
		if err != nil {
			return err
		}
		inst.windows = ws
	} else {
		windowStyle = inst.vp.View.WindowStyle
		if windowStyle != nil {
			height = windowStyle.Height
			width = windowStyle.Width
		}

		diagrams := inst.vp.View.Diagrams
		for i := range diagrams {
			ws, err := inst.setupAndBuildDiagrams(diagrams[i], windowStyle, width, height)
			if err != nil {
				return err
			}

			inst.windows = ws
		}
	}
	inst.renderer.Run()

	return nil
}

func (inst ProjectUIHolder) setupAndBuildMultiTabbedDiagramWindow(diagrams []viewproject.Diagram, width *float32, height *float32) ([]port.ExecutionWindow, error) {

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

func (inst ProjectUIHolder) setupAndBuildDiagramsMultiTabed(diagram viewproject.Diagram, width *float32, height *float32) (port.ExecutionWindow, error) {
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

func (inst ProjectUIHolder) setupAndBuildDiagrams(diagram viewproject.Diagram, windowStyle *viewproject.WindowStyle, width *float32, height *float32) ([]port.ExecutionWindow, error) {
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
func (inst ProjectUIHolder) buildDiagram(w port.ExecutionWindow, setup viewproject.SetupItem) port.ExecutionWindow {
	if setup.VariableType == viewproject.ValueTypeBoolean {

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
		})

		w.SetContent(boolFill)
		w.SetResizable(true)
	}
	return w
}

func (inst ProjectUIHolder) buildDiagramContent(setup viewproject.SetupItem) draw.Drawer {
	if setup.VariableType == viewproject.ValueTypeBoolean {

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
		})

		/*		content := container.NewMax(
				boolFill.Root(),
			)*/
		return boolFill

	}
	return nil
}

func (inst ProjectUIHolder) styleWindow(window port.ExecutionWindow, w *float32, h *float32) port.ExecutionWindow {

	if w != nil && h != nil {
		window.Resize(float32(*w), float32(*h))
	} else {
		window.Resize(inst.defaultWidth, inst.defaultHeight)
	}

	window.SetResizable(inst.defaultResizable)
	return window
}

func (inst ProjectUIHolder) styleTabbedWindow(window port.ExecutionWindow, w *float32, h *float32) port.ExecutionWindow {

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
}
