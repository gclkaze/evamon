package services

import (
	"fmt"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"

	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"

	"github.com/gclkaze/evamon/cmd/internal/ui/port"

	porter "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

type WidgetService struct {
	logger output.Printer
	setup  MainSetup

	renderer      port.Renderer
	drawerFactory porter.Factory
}

/*func NewWidgetService() *WidgetService {
	return &WidgetService{}
}*/

func NewWidgetService(r port.Renderer, df porter.Factory) *WidgetService {
	return &WidgetService{renderer: r, drawerFactory: df}
}

func (inst *WidgetService) SetSetup(setup MainSetup) {
	inst.setup = setup
	inst.logger = setup.GetPrinter()
}

/*
	func (inst *WidgetService) CreateWindow(vp *viewproject.ViewProject) {
		a := app.New()
		w := a.NewWindow(vp.View.Diagrams[0].Setup[0].Title)

		c, ok := colornames.Map["red"]
		if !ok {
			fmt.Println("color not found")
			return
		}
		//r, g, b, a := c.RGBA()
		bg := canvas.NewRectangle(c)

		content := container.NewMax(
			bg, // background layer
			container.NewCenter(
				widget.NewLabel(vp.View.Diagrams[0].Setup[0].Title),
			),
		)
		w.SetFixedSize(false)
		w.SetContent(content)
		w.Resize(fyne.NewSize(*vp.View.Diagrams[0].Setup[0].WindowStyle.Width, *vp.View.Diagrams[0].Setup[0].WindowStyle.Height)) // optional
		w.ShowAndRun()
	}
*/
func (inst *WidgetService) CreateWindow(vp *viewproject.ViewProject) error {
	if vp == nil || len(vp.View.Diagrams) == 0 || len(vp.View.Diagrams[0].Setup) == 0 {
		return fmt.Errorf("invalid view project: missing diagrams/setup")
	}

	setup := vp.View.Diagrams[0].Setup[0]
	title := setup.Title

	w, err := inst.renderer.NewExecutionWindow(title)
	if err != nil {
		return err
	}

	// Example bool fill drawer (last value only)
	red := colornames.Map["red"]
	green := colornames.Map["green"]

	boolFill := inst.drawerFactory.NewBoolFill(porter.BoolFillOptions{
		TrueColor:  green,
		FalseColor: red,
	})

	content := container.NewMax(
		boolFill.Root(),
		container.NewCenter(widget.NewLabel(title)),
	)

	w.SetResizable(true)
	w.SetContent(content)

	if setup.WindowStyle.Width != nil && setup.WindowStyle.Height != nil {
		w.Resize(float32(*setup.WindowStyle.Width), float32(*setup.WindowStyle.Height))
	}

	w.Show()
	return nil
}
