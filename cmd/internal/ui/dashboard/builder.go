package dashboardbuilder

import (
	"fmt"
	"strings"

	"github.com/gclkaze/evamon/cmd/internal/models"
	window "github.com/gclkaze/evamon/cmd/internal/ui/chart"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
	"github.com/magiconair/properties"

	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
)

// BindingTarget is what your socket binder needs:
// when a message arrives for (JobID, Variable), call Sink.Push(at, val).
type BindingTarget struct {
	JobID     string
	DiagramID string
	Variable  string
	Sink      dport.EvaWidget
}

// BuildResult is the output of the builder.
type BuildResult struct {
	Root     uport.UIObject
	Bindings []BindingTarget
}

// Builder builds a dashboard UI (tabs/rows/grids) and returns bindings for live data.
type Builder struct {
	Layout         uport.Layout
	Factory        dport.Factory // NewBoolFill / NewBarChart etc.
	ToolbarFactory dport.DiagramToolbarFactory
	chartRegistry  *port.ChartRegistry

	renderer       port.Renderer
	MinChartWidth  float32
	MinChartHeight float32

	MinChartBooleanWidth  float32
	MinChartBooleanHeight float32

	DefaultMaxPoints int
}

func New( /*layout, */ factory dport.Factory, toolbarFactory dport.DiagramToolbarFactory /*, inst.renderer.ChartRegistry()*/, renderer port.Renderer, props *properties.Properties) *Builder { //(layout uport.Layout, factory dport.Factory, toolbarFactory dport.DiagramToolbarFactory, chartRegistry *port.ChartRegistry) *Builder {

	/*	defaultWidth := props.GetFloat32("default_window_width", 400)
		defaultHeight := props.GetFloat32("default_window_height", 400)
		defaultResizable := props.GetBool("default_window_resizable", false)*/
	defaultMaxPoints := props.GetInt("default_barchart_maxpoints", 1000)

	return &Builder{
		Layout:         renderer.Layout(),
		Factory:        factory,
		ToolbarFactory: toolbarFactory,

		MinChartWidth:  400,
		MinChartHeight: 259,

		MinChartBooleanWidth:  100,
		MinChartBooleanHeight: 100,

		DefaultMaxPoints: defaultMaxPoints,
		chartRegistry:    renderer.ChartRegistry(),
		renderer:         renderer,
	}
}

// BuildDashboard builds the full dashboard as:
// Tabs( view ) -> VBox( rows... ) -> GridCols( columns... ) -> per-cell content
func (b *Builder) BuildDashboard(dp *vp.DashboardProject) (*BuildResult, error) {
	if dp == nil {
		return nil, fmt.Errorf("dashboard is nil")
	}
	if b.Layout == nil {
		return nil, fmt.Errorf("layout is nil")
	}
	if b.Factory == nil {
		return nil, fmt.Errorf("factory is nil")
	}

	var allBindings []BindingTarget
	var tabs []uport.TabItem

	for vi := range dp.Views {
		v := dp.Views[vi]

		viewObj, bindings, err := b.buildView(&v)
		if err != nil {
			return nil, fmt.Errorf("views[%d] (%q): %w", vi, v.Title, err)
		}
		allBindings = append(allBindings, bindings...)

		tabs = append(tabs, b.Layout.Tab(v.Title, viewObj))
	}

	root := b.Layout.Tabs(tabs...)

	root = b.Layout.VScroll(root)

	return &BuildResult{
		Root:     root,
		Bindings: allBindings,
	}, nil
}

func (b *Builder) buildDiagramForJob(jobID string, d *models.Diagram) (uport.UIObject, []BindingTarget, error) {
	if d == nil {
		return b.Layout.VBox(), nil, nil
	}
	if len(d.Setup) == 0 {
		return b.Layout.VBox(), nil, nil
	}

	var (
		content  uport.UIObject
		bindings []BindingTarget
		err      error
	)

	ref := port.NewDiagramUIRefs(d.ID, b.renderer)

	switch d.Type {
	case models.DiagramTypeBoolean:
		content, bindings, err = b.buildBooleanDiagram(jobID, d, ref)
	case models.DiagramTypeBar:
		content, bindings, err = b.buildBarDiagram(jobID, d, ref)
	case models.DiagramTypeLine:
		content, bindings, err = b.buildLineDiagram(jobID, d, ref)
	default:
		return nil, nil, fmt.Errorf("unsupported diagram type %q", d.Type)
	}

	if err != nil {
		return nil, nil, err
	}

	b.chartRegistry.Register(ref)

	toolbar := b.ToolbarFactory.Build(jobID, d)
	ref.Toolbar = toolbar
	ref.RegisterToolbar(toolbar)

	panel := b.wrapWithDiagramPanel(toolbar, content)
	ref.Panel = panel
	ref.ID = d.GetID()

	return panel, bindings, nil
}

func (b *Builder) buildView(v *vp.DashboardView) (uport.UIObject, []BindingTarget, error) {
	var rows []uport.UIObject
	var allBindings []BindingTarget

	for ri := range v.Rows {
		r := v.Rows[ri]

		rowObj, bindings, err := b.buildRow(&r)
		if err != nil {
			return nil, nil, fmt.Errorf("rows[%d]: %w", ri, err)
		}
		rows = append(rows, rowObj)
		allBindings = append(allBindings, bindings...)
	}

	return b.Layout.VBox(rows...), allBindings, nil
}

func (b *Builder) buildRow(r *vp.DashboardRow) (uport.UIObject, []BindingTarget, error) {
	if len(r.Columns) == 0 {
		// Should not happen if you validate, but keep safe.
		return b.Layout.VBox(), nil, nil
	}

	var cells []uport.UIObject
	var allBindings []BindingTarget

	for ci := range r.Columns {
		c := r.Columns[ci]

		cellObj, bindings, err := b.buildColumn(&c)
		if err != nil {
			return nil, nil, fmt.Errorf("columns[%d] (jobId=%q): %w", ci, c.JobID, err)
		}

		// Max makes each cell expand nicely within the grid.
		cells = append(cells, b.Layout.Max(cellObj))
		allBindings = append(allBindings, bindings...)
	}

	rowObj := b.Layout.GridCols(len(cells), cells...)
	return rowObj, allBindings, nil
}

func (b *Builder) wrapWithDiagramPanel(toolbar, content uport.UIObject) uport.UIObject {
	if toolbar == nil {
		return content
	}
	if content == nil {
		return toolbar
	}

	return b.Layout.Border(
		toolbar,
		nil,
		nil,
		nil,
		content,
	)
}

func (b *Builder) buildColumn(c *vp.DashboardColumn) (uport.UIObject, []BindingTarget, error) {
	// Column.View is a ViewWindow (reused struct) containing diagrams.
	diagrams := c.View.Diagrams
	if len(diagrams) == 0 {
		return b.Layout.VBox(), nil, nil
	}

	// If MultiTab: each diagram becomes a tab inside this cell.
	if c.View.MultiTab {
		var items []uport.TabItem
		var allBindings []BindingTarget

		for di := range diagrams {
			d := diagrams[di]

			obj, bindings, err := b.buildDiagramForJob(c.JobID, &d)
			if err != nil {
				return nil, nil, fmt.Errorf("diagrams[%d]: %w", di, err)
			}
			allBindings = append(allBindings, bindings...)

			title := defaultDiagramTitle(&d)
			items = append(items, b.Layout.Tab(title, obj))
		}

		return b.Layout.Tabs(items...), allBindings, nil
	}

	// Otherwise: stack diagrams vertically in the cell.
	var parts []uport.UIObject
	var allBindings []BindingTarget

	for di := range diagrams {
		d := diagrams[di]

		obj, bindings, err := b.buildDiagramForJob(c.JobID, &d)
		if err != nil {
			return nil, nil, fmt.Errorf("diagrams[%d]: %w", di, err)
		}
		parts = append(parts, obj)
		allBindings = append(allBindings, bindings...)
	}

	return b.Layout.VBox(parts...), allBindings, nil
}

// buildDiagramForJob creates UI objects for one Diagram and returns variable bindings.
// NOTE: a Diagram may have multiple setup items -> we usually create multiple widgets (one per setup item)
// and stack them.
func (b *Builder) buildDiagramForJobWithTitle(jobID string, d *models.Diagram) (uport.UIObject, []BindingTarget, error) {
	if d == nil {
		return b.Layout.VBox(), nil, nil
	}
	if len(d.Setup) == 0 {
		return b.Layout.VBox(), nil, nil
	}
	var (
		content  uport.UIObject
		bindings []BindingTarget
		err      error
	)
	ref := port.NewDiagramUIRefs(d.ID, b.renderer)

	switch d.Type {
	case models.DiagramTypeBoolean:
		content, bindings, err = b.buildBooleanDiagram(jobID, d, ref)
	case models.DiagramTypeBar:
		content, bindings, err = b.buildBarDiagram(jobID, d, ref)
	case models.DiagramTypeLine:
		content, bindings, err = b.buildLineDiagram(jobID, d, ref)
	default:
		return nil, nil, fmt.Errorf("unsupported diagram type %q", d.Type)
	}

	if err != nil {
		return nil, nil, err
	}
	panel := b.wrapWithDiagramTitle(jobID, d, content)
	ref.ID = d.GetID()
	ref.Panel = panel
	b.chartRegistry.Register(ref)

	return panel, bindings, nil
}

func (b *Builder) wrapWithDiagramTitle(jobID string, d *models.Diagram, content uport.UIObject) uport.UIObject {
	title := strings.TrimSpace(content.Title())
	if title == "" {
		// Good fallbacks: first setup title, first variable, or jobID/type
		if len(d.Setup) == 1 {
			// if your setup items have Title/Variable fields
			if t := strings.TrimSpace(d.Setup[0].Title); t != "" {
				title = t
			} else if v := strings.TrimSpace(d.Setup[0].Variable); v != "" {
				title = v
			}
		}
	}
	/*	if title == "" {
			title = jobID
		}
	*/
	// Card gives you a proper "titled box"
	return b.Layout.Card(title, "", content)
}

func (b *Builder) buildBooleanDiagram(jobID string, d *models.Diagram, ref *port.DiagramUIRefs) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	vars := window.CollectVariables(d)
	theItems := window.VariableStylesToLegendItems(vars)
	legendObj := b.Layout.DiagramLegend(theItems, nil)
	parts = append(parts, legendObj)

	ref.Legend = legendObj
	ref.RegisterLegend(legendObj)

	for si := range d.Setup {
		s := d.Setup[si]

		// ---- title per setup item (rendered via Layout, no Fyne imports here) ----
		titleText := strings.TrimSpace(s.Title)
		if titleText == "" {
			titleText = strings.TrimSpace(s.Variable)
		}
		if titleText != "" {
			parts = append(parts, b.Layout.Title(titleText))
		}
		w, err := window.CreateBoolDrawer(b.Factory, &s, si, b.MinChartBooleanWidth, b.MinChartBooleanHeight, d)
		if err != nil {
			return nil, nil, err
		}

		ref.Chart = w
		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:     jobID,
			DiagramID: d.ID,
			Variable:  s.Variable,
			Sink:      w,
		})

		if si < len(d.Setup)-1 {
			parts = append(parts, b.Layout.Separator())
		}
	}

	if dw, ok := ref.Chart.(dport.DiagramWidget); ok {
		parts = append(parts, dw.LastUpdatedLabel())
	}

	return b.Layout.VBox(parts...), bindings, nil
}

func (b *Builder) buildBarDiagram(jobID string, d *models.Diagram, ref *port.DiagramUIRefs) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	vars := window.CollectVariables(d)
	theItems := window.VariableStylesToLegendItems(vars)
	//parts = append(parts, legendObj)

	var w dport.DiagramWidget
	var err error
	for si := range d.Setup {
		s := d.Setup[si]

		w, err = window.CreateBarchartDrawer(b.Factory, &s, si, b.MinChartWidth, b.MinChartHeight, b.DefaultMaxPoints, d)
		if err != nil {
			return nil, nil, err
		}
		ref.Chart = w
		slot := b.Layout.Max(w)
		ref.ChartSlot = slot
		parts = append(parts, slot)
		bindings = append(bindings, BindingTarget{
			JobID:     jobID,
			DiagramID: d.ID,
			Variable:  s.Variable,
			Sink:      w,
		})
		ref.RegisterMainChart(w)
	}
	legendObj := b.Layout.DiagramLegend(theItems, func(it *port.LegendItem) {
		if len(theItems) == 1 {
			return
		}
		w.ToggleItem(it)
	})
	ref.Legend = legendObj
	ref.RegisterLegend(legendObj)
	parts = append([]uport.UIObject{legendObj}, parts...)
	parts = append(parts, w.LastUpdatedLabel())

	return b.Layout.VBox(parts...), bindings, nil
}

func (b *Builder) buildLineDiagram(jobID string, d *models.Diagram, ref *port.DiagramUIRefs) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	vars := window.CollectVariables(d)
	theItems := window.VariableStylesToLegendItems(vars)
	//legendObj := b.Layout.DiagramLegend(theItems, nil)
	//parts = append(parts, legendObj)
	var w dport.DiagramWidget
	var err error
	for si := range d.Setup {
		s := d.Setup[si]

		w, err = window.CreateLinechartDrawer(b.Factory, &s, si, b.MinChartWidth, b.MinChartHeight, b.DefaultMaxPoints, d)
		if err != nil {
			return nil, nil, err
		}
		ref.Chart = w
		slot := b.Layout.Max(w)
		ref.ChartSlot = slot
		parts = append(parts, slot)
		bindings = append(bindings, BindingTarget{
			JobID:     jobID,
			DiagramID: d.ID,
			Variable:  s.Variable,
			Sink:      w,
		})
		ref.RegisterMainChart(w)
	}

	legendObj := b.Layout.DiagramLegend(theItems, func(it *port.LegendItem) {
		if len(theItems) == 1 {
			return
		}
		w.ToggleItem(it)
	})
	ref.Legend = legendObj
	ref.RegisterLegend(legendObj)
	parts = append([]uport.UIObject{legendObj}, parts...)
	parts = append(parts, w.LastUpdatedLabel())

	return b.Layout.VBox(parts...), bindings, nil
}

func defaultDiagramTitle(d *models.Diagram) string {
	if d == nil {
		return ""
	}
	if len(d.Setup) > 0 && d.Setup[0].Title != "" {
		return d.Setup[0].Title
	}
	return string(d.Type)
}
