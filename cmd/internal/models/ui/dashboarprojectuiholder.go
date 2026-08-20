package ui

import (
	"sync"

	"fyne.io/fyne/v2/container"
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

func (inst *DashboardUIHolder) loadDefaults() {
	if inst.props == nil {
		return
	}

	inst.defaultWidth = inst.props.GetFloat32("default_window_width", 400)
	inst.defaultHeight = inst.props.GetFloat32("default_window_height", 400)
	inst.defaultResizable = inst.props.GetBool("default_window_resizable", false)
	inst.defaultMaxPoints = inst.props.GetInt("default_barchart_maxpoints", 1000)
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

	inst.loadDefaults()

	tf := ui.NewDiagramToolbarFactor(inst.renderer)
	builder := dashboardbuilder.New(inst.drawerFactory, tf, inst.renderer, inst.props)

	res, err := builder.BuildDashboard(dp)
	if err != nil {
		return err
	}

	content := res.Root
	if inst.logPanel != nil {
		splitObj := inst.renderer.Layout().VSplit(res.Root, inst.logPanel, 0.7)
		inst.configureLogPanel(splitObj)
		content = splitObj
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

// logPanelCfg is the subset of ExecutionLogPanel methods needed here.
type logPanelCfg interface {
	SetSplitCallbacks(onHide, onRestore func())
	SetNameResolver(func(string) string)
}

func (inst *DashboardUIHolder) configureLogPanel(splitObj port.UIObject) {
	cfg, ok := inst.logPanel.(logPanelCfg)
	if !ok {
		return
	}
	split, ok := splitObj.Native().(*container.Split)
	if !ok {
		return
	}
	const prevOffset = 0.7
	cfg.SetSplitCallbacks(
		func() { split.Offset = 1.0; split.Refresh() },
		func() { split.Offset = prevOffset; split.Refresh() },
	)
	cfg.SetNameResolver(buildDiagramNameResolver(inst.dp))
}

func buildDiagramNameResolver(dp *viewproject.DashboardProject) func(string) string {
	nameMap := make(map[string]string)
	for _, v := range dp.Views {
		for _, r := range v.Rows {
			for _, c := range r.Columns {
				for _, d := range c.View.Diagrams {
					if len(d.Setup) > 0 && d.Setup[0].Title != "" {
						nameMap[d.ID] = d.Setup[0].Title
					}
				}
			}
		}
	}
	return func(id string) string {
		if n, ok := nameMap[id]; ok {
			return n
		}
		if len(id) > 8 {
			return id[:8]
		}
		return id
	}
}
