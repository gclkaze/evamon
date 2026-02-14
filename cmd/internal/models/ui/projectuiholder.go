package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2/container"
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
	if inst.isMultitab {
		//for now, one window, multiple tabs
		//		inst.renderer.NewExecutionWindow(title string)
	} else {
		var windowStyle *viewproject.WindowStyle
		var height *float32
		var width *float32

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

			fmt.Printf("%d", len(ws))
		}
	}
	inst.renderer.Run()
	return nil
}

func (inst ProjectUIHolder) setupAndBuildDiagrams(diagram viewproject.Diagram, windowStyle *viewproject.WindowStyle, width *float32, height *float32) ([]port.ExecutionWindow, error) {
	//diagram := diagrams[i]
	theSetupItems := diagram.Setup
	var ws []port.ExecutionWindow
	for j := range theSetupItems {
		setup := theSetupItems[j]

		w, err := inst.renderer.NewExecutionWindow(setup.Title)
		if err != nil {
			return nil, err
		}

		if windowStyle == nil && setup.WindowStyle != nil {
			windowStyle = setup.WindowStyle
			width = windowStyle.Width
			height = windowStyle.Height
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

		content := container.NewMax(
			boolFill.Root(),
			/*			container.NewCenter(widget.NewLabel("ROFLMAO")),*/
		)

		w.SetContent(content)
		w.SetResizable(true)
	}
	return w
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
