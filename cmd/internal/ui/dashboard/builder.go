package dashboardbuilder

import (
	"fmt"
	"image/color"
	"strings"

	window "github.com/gclkaze/evamon/cmd/internal/ui/chart"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
	"golang.org/x/image/colornames"

	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
)

// BindingTarget is what your socket binder needs:
// when a message arrives for (JobID, Variable), call Sink.Push(at, val).
type BindingTarget struct {
	JobID    string
	Variable string
	Sink     dport.EvaWidget
}

// BuildResult is the output of the builder.
type BuildResult struct {
	Root     uport.UIObject
	Bindings []BindingTarget
}

// Builder builds a dashboard UI (tabs/rows/grids) and returns bindings for live data.
type Builder struct {
	Layout  uport.Layout
	Factory dport.Factory // NewBoolFill / NewBarChart etc.

	MinChartWidth  float32
	MinChartHeight float32

	MinChartBooleanWidth  float32
	MinChartBooleanHeight float32

	DefaultMaxPoints int
}

func New(layout uport.Layout, factory dport.Factory) *Builder {
	return &Builder{
		Layout:         layout,
		Factory:        factory,
		MinChartWidth:  400,
		MinChartHeight: 259,

		MinChartBooleanWidth:  100,
		MinChartBooleanHeight: 100,

		DefaultMaxPoints: 1000,
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
func (b *Builder) buildDiagramForJob(jobID string, d *vp.Diagram) (uport.UIObject, []BindingTarget, error) {
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
	switch d.Type {
	case vp.DiagramTypeBoolean:
		content, bindings, err = b.buildBooleanDiagram(jobID, d)
	case vp.DiagramTypeBar:
		content, bindings, err = b.buildBarDiagram(jobID, d)
	case vp.DiagramTypeLine:
		content, bindings, err = b.buildLineDiagram(jobID, d)
	default:
		return nil, nil, fmt.Errorf("unsupported diagram type %q", d.Type)
	}

	if err != nil {
		return nil, nil, err
	}

	return b.wrapWithDiagramTitle(jobID, d, content), bindings, nil
}

func (b *Builder) wrapWithDiagramTitle(jobID string, d *vp.Diagram, content uport.UIObject) uport.UIObject {
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

func (b *Builder) buildBooleanDiagram(jobID string, d *vp.Diagram) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	vars := b.collectVariables(d)
	theItems := variableStylesToLegendItems(vars)
	legendObj := b.Layout.DiagramLegend(theItems, nil)
	parts = append(parts, legendObj)

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
		w, err := window.CreateBoolDrawer(b.Factory, &s, si, b.MinChartBooleanWidth, b.MinChartBooleanHeight)
		if err != nil {
			return nil, nil, err
		}
		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:    jobID,
			Variable: s.Variable,
			Sink:     w,
		})

		if si < len(d.Setup)-1 {
			parts = append(parts, b.Layout.Separator())
		}
	}

	return b.Layout.VBox(parts...), bindings, nil
}

func (b *Builder) buildBarDiagram(jobID string, d *vp.Diagram) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	vars := b.collectVariables(d)
	theItems := variableStylesToLegendItems(vars)
	//parts = append(parts, legendObj)

	var w dport.DiagramWidget
	var err error
	for si := range d.Setup {
		s := d.Setup[si]

		w, err = window.CreateBarchartDrawer(b.Factory, &s, si, b.MinChartWidth, b.MinChartHeight, b.DefaultMaxPoints)
		if err != nil {
			return nil, nil, err
		}
		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:    jobID,
			Variable: s.Variable,
			Sink:     w,
		})
	}
	legendObj := b.Layout.DiagramLegend(theItems, func(it *port.LegendItem) {
		if len(theItems) == 1 {
			return
		}
		w.ToggleItem(it)
	})

	parts = append([]uport.UIObject{legendObj}, parts...)

	return b.Layout.VBox(parts...), bindings, nil
}

func (b *Builder) collectVariables(d *vp.Diagram) []dport.VariableStyle {
	var out []dport.VariableStyle

	for si := range d.Setup {
		s := &d.Setup[si]

		// ----------------------------
		// Multi-variable
		// ----------------------------
		if len(s.MultiVariableSetup) > 0 {
			for ci := range s.MultiVariableSetup {
				ms := &s.MultiVariableSetup[ci]

				var col color.Color
				txt := ""
				switch d.Type {
				case vp.DiagramTypeBar:
					if bs, ok := ms.DiagramStyle.(vp.BarStyle); ok {
						txt = bs.Bar
						col = lookupColor(bs.Bar)
					}

				case vp.DiagramTypeLine:
					if ls, ok := ms.DiagramStyle.(vp.LineStyle); ok {
						txt = ls.Line
						col = lookupColor(ls.Line)
					}
				}

				out = append(out, dport.VariableStyle{
					VariableName: ms.Variable,
					VarColor:     col,
					VarColorText: txt,
				})
			}
			continue
		}

		// ----------------------------
		// Single-variable
		// ----------------------------
		var col color.Color
		txt := ""
		switch d.Type {
		case vp.DiagramTypeBar:
			if bs, ok := s.DiagramStyle.(vp.BarStyle); ok {
				txt = bs.Axis
				col = lookupColor(bs.Axis)
			}

		case vp.DiagramTypeLine:
			if ls, ok := s.DiagramStyle.(vp.LineStyle); ok {
				txt = ls.Line
				col = lookupColor(ls.Line)
			}

		case vp.DiagramTypeBoolean:
			if bs, ok := s.DiagramStyle.(vp.BooleanStyle); ok {
				txt = bs.True
				col = lookupColor(bs.True)
			}
		}

		out = append(out, dport.VariableStyle{
			VariableName: s.Variable,
			VarColor:     col,
			VarColorText: txt,
		})
	}

	return out
}
func lookupColor(s string) color.Color {
	s = strings.ToLower(strings.TrimSpace(s))

	if c, ok := colornames.Map[s]; ok {
		return c
	}

	// fallback (neutral gray if unknown)
	return colornames.Gray
}

func variableStylesToLegendItems(vars []dport.VariableStyle) []port.LegendItem {
	out := make([]port.LegendItem, 0, len(vars))

	for i, v := range vars {
		out = append(out, port.LegendItem{
			Key:   v.VariableName, // stable id
			Label: v.VariableName, // you can change if you later add display name
			Color: v.VarColorText, // convert color.Color → string
			Index: i,
		})
	}

	return out
}

func (b *Builder) buildLineDiagram(jobID string, d *vp.Diagram) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	vars := b.collectVariables(d)
	theItems := variableStylesToLegendItems(vars)
	//legendObj := b.Layout.DiagramLegend(theItems, nil)
	//parts = append(parts, legendObj)
	var w dport.DiagramWidget
	var err error
	for si := range d.Setup {
		s := d.Setup[si]

		w, err = window.CreateLinechartDrawer(b.Factory, &s, si, b.MinChartWidth, b.MinChartHeight, b.DefaultMaxPoints)
		if err != nil {
			return nil, nil, err
		}

		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:    jobID,
			Variable: s.Variable,
			Sink:     w,
		})
	}

	legendObj := b.Layout.DiagramLegend(theItems, func(it *port.LegendItem) {
		if len(theItems) == 1 {
			return
		}
		w.ToggleItem(it)
	})
	parts = append([]uport.UIObject{legendObj}, parts...)

	return b.Layout.VBox(parts...), bindings, nil
}

func defaultDiagramTitle(d *vp.Diagram) string {
	if d == nil {
		return ""
	}
	if len(d.Setup) > 0 && d.Setup[0].Title != "" {
		return d.Setup[0].Title
	}
	return string(d.Type)
}
