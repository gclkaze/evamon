package services

import (
	"fmt"

	"github.com/gclkaze/evamon/cmd/internal/models/ui"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/magiconair/properties"

	"github.com/gclkaze/evamon/cmd/internal/ui/port"

	porter "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

type WidgetService struct {
	logger output.Printer
	setup  MainSetup

	renderer      port.Renderer
	drawerFactory porter.Factory
	//it holds all view projects -> ref -> map[Variable]-> multiple widgets
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

func (inst *WidgetService) CreateProjectUI(vp *viewproject.ViewProject, props *properties.Properties) error {
	if vp == nil || len(vp.View.Diagrams) == 0 || len(vp.View.Diagrams[0].Setup) == 0 {
		return fmt.Errorf("invalid view project: missing diagrams/setup")
	}
	uiholder := ui.NewProjectUIHolder(vp, inst.renderer, inst.drawerFactory, props)
	err := uiholder.Create()
	if err != nil {
		return err
	}
	return nil
}

func (inst *WidgetService) Update(vp *viewproject.ViewProject) error {
	return nil
}
