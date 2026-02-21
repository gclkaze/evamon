package ui

import (
	"sync"

	dashboardbuilder "github.com/gclkaze/evamon/cmd/internal/ui/dashboard"
	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/magiconair/properties"
)

type DashboardUIHolder struct {
	dp            *viewproject.DashboardProject
	window        port.ExecutionWindow
	renderer      port.Renderer
	drawerFactory draw.Factory

	props *properties.Properties

	defaultWidth     float32
	defaultHeight    float32
	defaultResizable bool
	defaultMaxPoints int

	jobRouter *JobRouter

	mu          sync.RWMutex
	unsubscribe map[string]map[draw.EvaWidget]func()
}

func NewDashboardUIHolder(dp *viewproject.DashboardProject, renderer port.Renderer, drawerFactory draw.Factory, props *properties.Properties, jobRouter *JobRouter) *DashboardUIHolder {
	return &DashboardUIHolder{dp: dp, renderer: renderer, props: props, drawerFactory: drawerFactory, jobRouter: jobRouter, unsubscribe: make(map[string]map[draw.EvaWidget]func())}
}

func (inst *DashboardUIHolder) SetOnClosed(close func()) {
	inst.window.SetOnClosed(close)
}
func (inst *DashboardUIHolder) Create(dp *viewproject.DashboardProject) error {

	win, err := inst.renderer.NewExecutionWindow(inst.dp.Title)

	if err != nil {
		return err
	}
	layout := inst.renderer.Layout()
	builder := dashboardbuilder.New(layout, inst.drawerFactory)

	res, err := builder.BuildDashboard(dp)
	if err != nil {
		return err
	}

	win.SetContent(res.Root)
	win.Resize(1100, 700)
	win.Show()

	// Bindings: subscribe to sockets
	// for each incoming message (jobId, variable, at, value) -> route:
	for _, b := range res.Bindings {
		inst.jobRouter.Register(b.JobID, b.Variable, b.Sink)
	}

	inst.window = win

	return nil
}
func (inst *DashboardUIHolder) Run() {
	inst.renderer.Run()
}
