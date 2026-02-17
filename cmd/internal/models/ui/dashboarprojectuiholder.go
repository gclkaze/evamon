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
func (inst *DashboardUIHolder) Create() error {

	win, _ := inst.renderer.NewExecutionWindow(inst.dp.Title)

	layout := inst.renderer.Layout()
	builder := dashboardbuilder.New(layout, inst.drawerFactory)

	dp, _ := viewproject.LoadDashboardProject("dashboard.json")
	res, err := builder.BuildDashboard(dp)
	if err != nil {
		panic(err)
	}

	win.SetContent(res.Root)
	win.Show()

	// Bindings: subscribe to sockets
	// for each incoming message (jobId, variable, at, value) -> route:
	for _, b := range res.Bindings {
		inst.jobRouter.Register(b.JobID, b.Variable, b.Sink)
	}

	return nil
}
func (inst *DashboardUIHolder) Run() {
	inst.renderer.Run()
}

/*

type Router struct {
	mu sync.RWMutex
	m  map[string]map[string][]port.EvaWidget // jobID -> variable -> sinks
}

func NewRouter() *Router {
	return &Router{m: make(map[string]map[string][]port.EvaWidget)}
}

func (r *Router) Register(jobID, variable string, sink port.EvaWidget) {
	r.mu.Lock()
	defer r.mu.Unlock()

	vm, ok := r.m[jobID]
	if !ok {
		vm = make(map[string][]port.EvaWidget)
		r.m[jobID] = vm
	}
	vm[variable] = append(vm[variable], sink)
}

func (r *Router) Push(jobID, variable string, at time.Time, val any) {
	r.mu.RLock()
	sinks := r.m[jobID][variable]
	r.mu.RUnlock()

	for _, s := range sinks {
		s.Push(at, val)
	}
}

*/
