package ui

import (
	"github.com/gclkaze/evamon/cmd/internal/models"
	dport "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

// ToolbarSelector picks the appropriate per-type toolbar factory and delegates to it.
// It implements dport.DiagramToolbarFactory.
type ToolbarSelector struct {
	chart   *ChartToolbarFactory
	boolFil *BoolFillToolbarFactory
}

func NewDiagramToolbarFactor(renderer port.Renderer) *ToolbarSelector {
	return &ToolbarSelector{
		chart:   &ChartToolbarFactory{renderer: renderer},
		boolFil: &BoolFillToolbarFactory{},
	}
}

func (s *ToolbarSelector) Build(jobID string, d port.IDiagram) port.UIObject {
	if d == nil {
		return nil
	}
	switch d.GetType() {
	case models.DiagramTypeBoolean:
		return s.boolFil.Build(jobID, d)
	default:
		return s.chart.Build(jobID, d)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ChartToolbarFactory — full toolbar for bar and line charts.
// ──────────────────────────────────────────────────────────────────────────────

type ChartToolbarFactory struct {
	renderer port.Renderer
}

func (f *ChartToolbarFactory) Build(jobID string, d port.IDiagram) port.UIObject {
	toolbar := f.renderer.Layout().HBox(f.buildChildren(jobID, d)...)

	filterRow, filterRowInner := f.buildFilterRow(jobID, d)

	if refs, ok := f.renderer.ChartRegistry().Get(d.GetID()); ok {
		refs.RegisterToolbar(toolbar)
		refs.SetRebuildWrapper(func() {
			newToolbarChildren := f.buildChildren(jobID, d)
			f.renderer.Layout().ReplaceHBoxContent(toolbar, newToolbarChildren...)

			newFilterChildren := f.buildFilterRowChildren(jobID, d)
			if filterRow != nil {
				f.renderer.Layout().ReplaceHBoxContent(filterRowInner, newFilterChildren...)
			}
		})
	}

	if filterRow != nil {
		return f.renderer.Layout().VBox(toolbar, filterRow)
	}
	return toolbar
}

func (f *ChartToolbarFactory) buildChildren(jobID string, d port.IDiagram) []port.UIObject {
	title := diagramTitle(d)
	setupItem := primarySetupItem(d)
	theRef, exists := f.renderer.ChartRegistry().Get(d.GetID())

	children := []port.UIObject{f.renderer.Layout().Title(title), f.renderer.Layout().Spacer()}

	if setupItem != nil && setupItem.HasFilter() {
		filterToggle := f.renderer.Controls().Check("filters", setupItem.IsFilterEnabled(), func(v bool) {
			f.renderer.Actions().SetFilterEnabled(jobID, d, v)
		})
		children = append(children, filterToggle)
		if exists {
			theRef.RegisterFilterEnabledCheckbox(filterToggle)
		}
	}

	filtersBtn := f.renderer.Controls().IconButton(port.IconFilters, func() {
		f.renderer.Actions().Filters(jobID, d)
	})
	if exists {
		theRef.RegisterFilterButton(filtersBtn)
	}

	downloadBtn := f.renderer.Controls().IconMenu(port.IconDownload, []port.MenuItem{
		{Label: "JSON", Action: func() { f.renderer.Actions().DownloadJSON(jobID, d) }},
		{Label: "CSV", Action: func() { f.renderer.Actions().DownloadCSV(jobID, d) }},
	})
	if exists {
		theRef.RegisterDownloadButton(downloadBtn)
	}

	snapshotBtn := f.renderer.Controls().IconButton(port.IconSnapshot, func() {
		f.renderer.Actions().SnapshotChart(jobID, d)
	})

	maximizeBtn := f.renderer.Controls().IconButton(port.IconMaximize, func() {
		f.renderer.Actions().Maximize(jobID, d)
	})
	if exists {
		theRef.RegisterMaximizeButton(maximizeBtn)
	}

	var zoomInBtn, zoomOutBtn port.UIObject
	if exists {
		if zoomable, ok := theRef.Chart.(dport.DiagramWidget); ok {
			zoomInBtn = f.renderer.Controls().IconButton(port.IconZoomIn, func() {
				zoomable.ZoomIn()
			})
			zoomOutBtn = f.renderer.Controls().IconButton(port.IconZoomOut, func() {
				zoomable.ZoomOut()
			})
		}
	}

	operationsBtn := f.renderer.Controls().IconButton(port.IconOperations, func() {
		f.renderer.Actions().Operations(jobID, d)
	})

	children = append(children, filtersBtn)
	if zoomInBtn != nil {
		children = append(children, zoomInBtn, zoomOutBtn)
	}
	children = append(children, downloadBtn, snapshotBtn, operationsBtn, maximizeBtn)

	return children
}

func (f *ChartToolbarFactory) buildFilterRowChildren(jobID string, d port.IDiagram) []port.UIObject {
	setupItem := primarySetupItem(d)
	if setupItem == nil || !setupItem.HasFilter() {
		return nil
	}
	filter := setupItem.Filter
	if filter == nil || filter.Setup == nil || len(filter.Setup.Components) == 0 {
		return nil
	}

	var children []port.UIObject
	for i := range filter.Setup.Components {
		c := filter.Setup.Components[i]
		label := models.GetFilterLabel(c.Label, c.Expression, i)
		check := f.renderer.Controls().Check(label, c.Enabled, func(v bool) {
			f.renderer.Actions().SetFilterComponentEnabled(jobID, d, c.ID, v)
		})
		children = append(children, check)
	}
	return children
}

func (f *ChartToolbarFactory) buildFilterRow(jobID string, d port.IDiagram) (port.UIObject, port.UIObject) {
	children := f.buildFilterRowChildren(jobID, d)
	hbox := f.renderer.Layout().HBox(children...)
	scroll := f.renderer.Layout().HScroll(hbox)
	return scroll, hbox
}

// ──────────────────────────────────────────────────────────────────────────────
// BoolFillToolbarFactory — no toolbar for boolean diagrams.
// ──────────────────────────────────────────────────────────────────────────────

type BoolFillToolbarFactory struct{}

func (f *BoolFillToolbarFactory) Build(_ string, _ port.IDiagram) port.UIObject {
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Shared helpers
// ──────────────────────────────────────────────────────────────────────────────

func primarySetupItem(d port.IDiagram) *models.SetupItem {
	setups := d.GetSetup()
	if d == nil || len(setups) == 0 {
		return nil
	}
	return &setups[0]
}

func diagramTitle(d port.IDiagram) string {
	if d == nil {
		return ""
	}
	setups := d.GetSetup()
	if len(setups) == 0 {
		return ""
	}
	if setups[0].Title != "" {
		return setups[0].Title
	}
	return string(d.GetType())
}
