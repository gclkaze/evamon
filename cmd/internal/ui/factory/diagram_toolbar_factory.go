package ui

import (
	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type DiagramToolbarFactory struct {
	renderer port.Renderer
}

func primarySetupItem(d port.IDiagram) *models.SetupItem {
	setups := d.GetSetup()
	if d == nil || len(setups) == 0 {
		return nil
	}
	return &setups[0]
}

func NewDiagramToolbarFactor(renderer port.Renderer) *DiagramToolbarFactory {
	return &DiagramToolbarFactory{renderer: renderer}
}

func (f *DiagramToolbarFactory) buildChildren(jobID string, d port.IDiagram) []port.UIObject {
	title := diagramTitle(d)
	setupItem := primarySetupItem(d)
	theRef, exists := f.renderer.ChartRegistry().Get(d.GetID())

	var children []port.UIObject
	children = append(children, f.renderer.Layout().Title(title), f.renderer.Layout().Spacer())

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
		theRef.RegisterDownloadButton(filtersBtn)
	}
	maximizeBtn := f.renderer.Controls().IconButton(port.IconMaximize, func() {
		f.renderer.Actions().Maximize(jobID, d)
	})
	if exists {
		theRef.RegisterMaximizeButton(filtersBtn)
	}
	children = append(children, filtersBtn, downloadBtn, maximizeBtn)
	return children
}

func (f *DiagramToolbarFactory) Build(jobID string, d port.IDiagram) port.UIObject {
	if d == nil {
		return nil
	}

	children := f.buildChildren(jobID, d)
	toolbar := f.renderer.Layout().HBox(children...)

	if refs, ok := f.renderer.ChartRegistry().Get(d.GetID()); ok {
		refs.RegisterToolbar(toolbar)
		refs.SetRebuildToolbar(func() []port.UIObject {
			return f.buildChildren(jobID, d)
		})
	}

	return toolbar
}

func diagramTitle(d port.IDiagram) string {
	if d == nil {
		return ""
	}
	setups := d.GetSetup()
	if setups == nil {
		return ""
	}

	if len(setups) == 0 {
		return ""
	}

	if setups[0].Title != "" {
		return setups[0].Title
	}

	return string(d.GetType())
}
