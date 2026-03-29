package ui

import (
	"sync"

	dashboardbuilder "github.com/gclkaze/evamon/cmd/internal/ui/dashboard"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	draw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	ui "github.com/gclkaze/evamon/cmd/internal/ui/factory"
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

	jobRouter            *JobRouter
	triggerSenderFactory func(jobID, diagramID string) data.TriggerSendFunc
	logPanel             port.UIObject // optional; when set a VSplit is added below the dashboard

	mu          sync.RWMutex
	unsubscribe map[string]map[draw.EvaWidget]func()
}

func NewDashboardUIHolder(dp *viewproject.DashboardProject, renderer port.Renderer, drawerFactory draw.Factory, props *properties.Properties, jobRouter *JobRouter) *DashboardUIHolder {
	return &DashboardUIHolder{dp: dp, renderer: renderer, props: props, drawerFactory: drawerFactory, jobRouter: jobRouter, unsubscribe: make(map[string]map[draw.EvaWidget]func())}
}

func (inst *DashboardUIHolder) SetTriggerSenderFactory(fn func(jobID, diagramID string) data.TriggerSendFunc) {
	inst.triggerSenderFactory = fn
}

// SetLogPanel attaches an optional log panel that is shown below the diagram grid
// via a resizable VSplit (70% diagrams / 30% logs).
func (inst *DashboardUIHolder) SetLogPanel(p port.UIObject) {
	inst.logPanel = p
}

func (inst *DashboardUIHolder) SetOnClosed(close func()) {
	inst.window.SetOnClosed(close)
}
func (inst *DashboardUIHolder) Create(dp *viewproject.DashboardProject) error {

	win, err := inst.renderer.NewExecutionWindow(inst.dp.Title)

	if err != nil {
		return err
	}
	tf := ui.NewDiagramToolbarFactor(inst.renderer)
	builder := dashboardbuilder.New(inst.drawerFactory, tf, inst.renderer)

	res, err := builder.BuildDashboard(dp)
	if err != nil {
		return err
	}

	content := res.Root
	if inst.logPanel != nil {
		content = inst.renderer.Layout().VSplit(res.Root, inst.logPanel, 0.7)
	}
	win.SetContent(content)
	win.Resize(1100, 700)
	win.Show()

	// Bindings: subscribe to sockets
	// for each incoming message (jobId, variable, at, value) -> route:
	seen := make(map[string]struct{})
	for _, b := range res.Bindings {
		inst.jobRouter.Register(b.JobID, b.Variable, b.Sink)
		if inst.triggerSenderFactory != nil && b.DiagramID != "" {
			if _, already := seen[b.DiagramID]; !already {
				seen[b.DiagramID] = struct{}{}
				if dw, ok := b.Sink.(draw.DiagramWidget); ok {
					ds := dw.GetDataSeries()
					if ds != nil {
						ds.SetTriggerSender(inst.triggerSenderFactory(b.JobID, b.DiagramID))

					}
				}
			}
		}
	}

	inst.window = win

	return nil
}
func (inst *DashboardUIHolder) Run() {
	inst.renderer.Run()
}
