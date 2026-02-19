package services

import (
	"fmt"
	"time"

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
	variableContainer *ui.VariableContainer
	jobRouter         *ui.JobRouter

	uiHolder        *ui.ProjectUIHolder
	dashboardHolder *ui.DashboardUIHolder
}

func NewWidgetService(r port.Renderer, df porter.Factory) *WidgetService {
	return &WidgetService{renderer: r, drawerFactory: df, variableContainer: ui.NewVariableContainer(), jobRouter: ui.NewJobRouter()}
}

func (inst *WidgetService) SetSetup(setup MainSetup) {
	inst.setup = setup
	inst.logger = setup.GetPrinter()
}

func (inst *WidgetService) SetOnClosed(close func()) {
	if inst.uiHolder != nil {
		inst.uiHolder.SetOnClosed(close)
	}

	if inst.dashboardHolder != nil {
		inst.dashboardHolder.SetOnClosed(close)
	}
}

func (inst *WidgetService) CreateProjectUI(vp *viewproject.ViewProject, props *properties.Properties) error {
	if vp == nil || len(vp.View.Diagrams) == 0 || len(vp.View.Diagrams[0].Setup) == 0 {
		return fmt.Errorf("invalid view project: missing diagrams/setup")
	}
	inst.uiHolder = ui.NewProjectUIHolder(vp, inst.renderer, inst.drawerFactory, props, inst.variableContainer)
	err := inst.uiHolder.Create()
	if err != nil {
		return err
	}
	return nil
}

func (inst *WidgetService) CreateDashboardProjectUI(dp *viewproject.DashboardProject, props *properties.Properties) error {
	if dp == nil || dp.IsEmpty() {
		return fmt.Errorf("invalid dashboard project: the project is empty")
	}
	inst.dashboardHolder = ui.NewDashboardUIHolder(dp, inst.renderer, inst.drawerFactory, props, inst.jobRouter)
	err := inst.dashboardHolder.Create(dp)
	if err != nil {
		return err
	}
	return nil
}

func (inst *WidgetService) Update(vp *viewproject.ViewProject) error {
	return nil
}

func (inst *WidgetService) Run() {
	if inst.uiHolder != nil {
		inst.uiHolder.Run()
	}

	if inst.dashboardHolder != nil {
		inst.dashboardHolder.Run()
	}
}

func (inst *WidgetService) DispatchValue(jobID string, varName string, t time.Time, value any) {
	if inst.jobRouter != nil {
		inst.jobRouter.Push(jobID, varName, t, value)
	}
	inst.variableContainer.Push(varName, t, value)
}
