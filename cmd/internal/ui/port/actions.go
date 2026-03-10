package port

import vp "github.com/gclkaze/evamon/cmd/internal/viewproject"

type DiagramActions interface {
	Maximize(jobID string, d *vp.Diagram)
	DownloadJSON(jobID string, d *vp.Diagram)
	DownloadCSV(jobID string, d *vp.Diagram)
	Filters(jobID string, d *vp.Diagram)
	SetFilterEnabled(jobID string, d *vp.Diagram, enabled bool)
}
