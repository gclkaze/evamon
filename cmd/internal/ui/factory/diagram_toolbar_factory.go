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
	filterRow := f.buildFilterRow(jobID, d)

	if refs, ok := f.renderer.ChartRegistry().Get(d.GetID()); ok {
		refs.RegisterToolbar(toolbar)
		refs.SetRebuildToolbar(func() []port.UIObject {
			return f.buildChildren(jobID, d)
		})
		/*		if filterRow != nil {
				refs.RegisterFilterRow(filterRow)
				refs.SetRebuildFilterRow(func() port.UIObject {
					return f.buildFilterRowChildren(jobID, d)
				})
			}*/

		if filterRow != nil {
			refs.RegisterFilterRow(filterRow)
			refs.SetRebuildFilterRow(func() port.UIObject {
				return f.buildFilterRow(jobID, d)
			})
		}

		refs.SetRebuildWrapper(func() {
			// rebuild toolbar in place
			newToolbarChildren := f.buildChildren(jobID, d)
			f.renderer.Layout().ReplaceHBoxContent(toolbar, newToolbarChildren...)

			// rebuild filter row in place
			newFilterChildren := f.buildFilterRowChildren(jobID, d)
			if filterRow != nil {
				f.renderer.Layout().ReplaceHBoxContent(filterRow, newFilterChildren...)
			}
		})
	}

	if filterRow != nil {
		return f.renderer.Layout().VBox(toolbar, filterRow)
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
func (f *DiagramToolbarFactory) buildFilterRowChildren(jobID string, d port.IDiagram) []port.UIObject {
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

func (f *DiagramToolbarFactory) buildFilterRow(jobID string, d port.IDiagram) port.UIObject {
	children := f.buildFilterRowChildren(jobID, d)
	if len(children) == 0 {
		return nil
	}
	return f.renderer.Layout().HBox(children...)
}
