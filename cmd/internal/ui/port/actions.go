package port

type DiagramActions interface {
	Maximize(jobID string, d IDiagram)
	DownloadJSON(jobID string, d IDiagram)
	DownloadCSV(jobID string, d IDiagram)
	SnapshotChart(jobID string, d IDiagram)
	Filters(jobID string, d IDiagram)
	Operations(jobID string, d IDiagram)
	SetFilterEnabled(jobID string, d IDiagram, enabled bool)
	SetFilterComponentEnabled(jobID string, d IDiagram, componentID string, enabled bool)
}
