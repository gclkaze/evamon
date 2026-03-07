package fynerenderer

import (
	"log"

	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
)

type DiagramActionHandler struct {
	// add dependencies later:
	// WindowManager
	// ExportService
	// FilterService
}

func (h DiagramActionHandler) Maximize(jobID string, d *vp.Diagram) {
	log.Printf("maximize clicked for job=%s diagram=%s\n", jobID, d.GetName())
}

func (h DiagramActionHandler) DownloadJSON(jobID string, d *vp.Diagram) {
	log.Printf("download JSON clicked for job=%s diagram=%s\n", jobID, d.GetName())
}

func (h DiagramActionHandler) DownloadCSV(jobID string, d *vp.Diagram) {
	log.Printf("download CSV clicked for job=%s diagram=%s\n", jobID, d.GetName())
}

func (h DiagramActionHandler) Filters(jobID string, d *vp.Diagram) {
	log.Printf("filters clicked for job=%s diagram=%s\n", jobID, d.GetName())
}
