package ui

/*import (
	vp "yourmodule/path/to/vp"
	dport "yourmodule/ui/diagrams/port"
	uport "yourmodule/ui/port"
)*/

import (
	uport "github.com/gclkaze/evamon/cmd/internal/ui/port"
	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
)

type DefaultDiagramToolbarFactory struct {
	Layout   uport.Layout
	Controls uport.Controls
	Actions  uport.DiagramActions
}

func primarySetupItem(d *vp.Diagram) *vp.SetupItem {
	if d == nil || len(d.Setup) == 0 {
		return nil
	}
	return &d.Setup[0]
}

func (f *DefaultDiagramToolbarFactory) Build(jobID string, d *vp.Diagram) uport.UIObject {
	if d == nil {
		return nil
	}

	title := diagramTitle(d)
	setupItem := primarySetupItem(d)

	var children []uport.UIObject
	children = append(children, f.Layout.Title(title), f.Layout.Spacer())

	if setupItem != nil && setupItem.HasFilter() {
		filterToggle := f.Controls.Check("filters", setupItem.IsFilterEnabled(), func(v bool) {
			f.Actions.SetFilterEnabled(jobID, d, v)
		})
		children = append(children, filterToggle)
	}

	filtersBtn := f.Controls.IconButton(uport.IconFilters, func() {
		f.Actions.Filters(jobID, d)
	})

	downloadBtn := f.Controls.IconMenu(uport.IconDownload, []uport.MenuItem{
		{
			Label: "JSON",
			Action: func() {
				f.Actions.DownloadJSON(jobID, d)
			},
		},
		{
			Label: "CSV",
			Action: func() {
				f.Actions.DownloadCSV(jobID, d)
			},
		},
	})

	maximizeBtn := f.Controls.IconButton(uport.IconMaximize, func() {
		f.Actions.Maximize(jobID, d)
	})

	children = append(children, filtersBtn, downloadBtn, maximizeBtn)

	return f.Layout.HBox(children...)
}

func diagramTitle(d *vp.Diagram) string {
	if d == nil {
		return ""
	}

	if d.Setup == nil {
		return ""
	}

	if len(d.Setup) == 0 {
		return ""
	}

	if d.Setup[0].Title != "" {
		return d.Setup[0].Title
	}

	return string(d.Type)
}
