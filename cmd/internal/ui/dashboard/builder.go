package dashboardbuilder

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
	"golang.org/x/image/colornames"

	"github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
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
}

func New(layout uport.Layout, factory dport.Factory) *Builder {
	return &Builder{
		Layout:         layout,
		Factory:        factory,
		MinChartWidth:  400,
		MinChartHeight: 259,

		MinChartBooleanWidth:  100,
		MinChartBooleanHeight: 100,
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

		// Extract style if present
		var trueColor, falseColor color.RGBA
		if s.DiagramStyle != nil {
			st, ok := s.DiagramStyle.(vp.BooleanStyle)
			if !ok {
				return nil, nil, fmt.Errorf("setup[%d]: diagramStyle is not BooleanStyle", si)
			}
			trueColor = colornames.Map[st.True]
			falseColor = colornames.Map[st.False]
		} else {
			falseColor = colornames.Map["red"]
			trueColor = colornames.Map["green"]
		}

		// Build widget; since we render title outside, avoid duplicating it inside the widget
		w := b.Factory.NewBoolFill(dport.BoolFillOptions{
			TrueColor:  trueColor,
			FalseColor: falseColor,
		}, "", s.Description, b.MinChartBooleanWidth, b.MinChartBooleanHeight)

		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:    jobID,
			Variable: s.Variable,
			Sink:     w,
		})

		// Optional: visual spacing between setup blocks
		if si < len(d.Setup)-1 {
			parts = append(parts, b.Layout.Separator())
		}
	}

	return b.Layout.VBox(parts...), bindings, nil
}

func (b *Builder) buildBarDiagram(jobID string, d *vp.Diagram) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	for si := range d.Setup {
		s := d.Setup[si]

		var axisColor, backgroundColor color.RGBA

		if s.DiagramStyle != nil {
			if bs, ok := s.DiagramStyle.(vp.BarStyle); ok {
				axisColor = colornames.Map[bs.Axis]
				backgroundColor = colornames.Map[bs.Background]
			}
		} else {
			axisColor = colornames.Map["red"]
			backgroundColor = colornames.Map["black"]

		}

		w := b.Factory.NewBarChart(dport.BarChartOptions{
			// If you have per-setup windowStyle, you can map it here too.
			// Width/Height are typically 0 so layout controls the size.
			MaxPoints:  0, // or set from somewhere else (global default)
			Axis:       axisColor,
			Background: backgroundColor,
		}, s.Title, s.Description, b.MinChartWidth, b.MinChartHeight)

		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:    jobID,
			Variable: s.Variable,
			Sink:     w,
		})
	}

	return b.Layout.VBox(parts...), bindings, nil
}

func (b *Builder) buildLineDiagram(jobID string, d *vp.Diagram) (uport.UIObject, []BindingTarget, error) {
	var parts []uport.UIObject
	var bindings []BindingTarget

	for si := range d.Setup {
		s := d.Setup[si]
		opts := port.DefaultLineChartOptions()

		if s.DiagramStyle != nil {
			if bs, ok := s.DiagramStyle.(vp.LineStyle); ok {
				lineColor := colornames.Map[bs.Line]
				backgroundColor := colornames.Map[bs.Background]

				opts.Line = lineColor
				opts.Background = backgroundColor
			}
		}

		w := b.Factory.NewLineChart(
			opts,
			s.Title,
			s.Description,
			800, 260, // initial raster (so it doesn’t start tiny/blurry)
			b.MinChartWidth, b.MinChartHeight, // widget hint; layout will expand it
		)

		parts = append(parts, w)
		bindings = append(bindings, BindingTarget{
			JobID:    jobID,
			Variable: s.Variable,
			Sink:     w,
		})
	}

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

// (Optional) Useful if your binder wants to push "now" without parsing timestamps.
func pushNow(w dport.EvaWidget, val any) {
	w.Push(time.Now(), val)
}
