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

func (f *DefaultDiagramToolbarFactory) Build(jobID string, d *vp.Diagram) uport.UIObject {
	if d == nil {
		return nil
	}

	title := diagramTitle(d)

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

	return f.Layout.HBox(
		f.Layout.Title(title),
		f.Layout.Spacer(),
		filtersBtn,
		downloadBtn,
		maximizeBtn,
	)
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
