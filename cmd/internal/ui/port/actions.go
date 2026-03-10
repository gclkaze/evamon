package port

type DiagramActions interface {
	Maximize(jobID string, d IDiagram)
	DownloadJSON(jobID string, d IDiagram)
	DownloadCSV(jobID string, d IDiagram)
	Filters(jobID string, d IDiagram)
	SetFilterEnabled(jobID string, d IDiagram, enabled bool)
}
