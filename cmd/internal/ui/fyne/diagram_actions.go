package fynerenderer

import (
	"fmt"
	"log"
	"strings"

	"github.com/gclkaze/evamon/cmd/internal/models"
	diaw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"

	"github.com/gclkaze/evamon/cmd/internal/ui/port"
	dia "github.com/gclkaze/evamon/cmd/internal/ui/port"
	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/gclkaze/evamon/pkg/utils"
)

type DiagramActionHandler struct {
	// add dependencies later:
	// WindowManager
	// ExportService
	// Filter validator / parser
	// Project saver
	ChartRegistry *dia.ChartRegistry

	snapshotComponents []models.FilterComponent
	snaphshotMode      models.FilterMode
}

func (h *DiagramActionHandler) Maximize(jobID string, d port.IDiagram) {
	log.Printf("maximize clicked for job=%s diagram=%s\n", jobID, d.GetName())
}

func (h *DiagramActionHandler) DownloadJSON(jobID string, d port.IDiagram) {
	log.Printf("download JSON clicked for job=%s diagram=%s\n", jobID, d.GetName())
}

func (h *DiagramActionHandler) DownloadCSV(jobID string, d port.IDiagram) {
	log.Printf("download CSV clicked for job=%s diagram=%s\n", jobID, d.GetName())
}

func (h *DiagramActionHandler) RefreshDiagram(d port.IDiagram) {
	if x, ok := h.ChartRegistry.Get(d.GetID()); ok {
		if c, ok := x.GetMainChart(); ok {
			c.Refresh()
		}
	}
}

func (h *DiagramActionHandler) SetFilterEnabled(jobID string, d port.IDiagram, enabled bool) {
	setups := d.GetSetup()

	if d == nil || len(setups) == 0 {
		return
	}

	setupItem := &setups[0]
	if setupItem.Filter == nil {
		return
	}

	setupItem.Filter.Enabled = enabled

	h.RefreshDiagram(d)

	owner := d.GetDiagramOwner()
	switch p := owner.(type) {
	case *vp.ViewProject:
		p.Save()
	case *vp.DashboardProject:
		p.Save()
	default:

	}

	// TODO: persist project/config if needed
	// TODO: trigger chart refresh/re-filter if needed

	log.Printf(
		"set filter enabled=%v for job=%s diagram=%s\n",
		enabled, jobID, d.GetName(),
	)
}

func (h *DiagramActionHandler) Filters(jobID string, d port.IDiagram) {
	log.Printf("filters clicked for job=%s diagram=%s\n", jobID, d.GetName())
	setups := d.GetSetup()
	if d == nil || len(setups) == 0 {
		return
	}

	win := NewChildWindow("Filters", 520, 360)
	setupItem := &setups[0]

	// snapshot before editor opens
	h.snapshotComponents = nil
	if setupItem.Filter != nil && setupItem.Filter.Setup != nil {
		h.snapshotComponents = append([]models.FilterComponent(nil), setupItem.Filter.Setup.Components...)
		h.snaphshotMode = setupItem.GetFilterMode()
	}

	editor := NewFilterEditor(
		"Filters - "+d.GetName(),
		setupItem.Filter,
		d,
		func(expr string) error {
			return h.validateExpression(expr, d)
		},
		func(f *models.Filter, d dia.IDiagram) error {
			if err := h.saveFilters(jobID, setupItem, f, d); err != nil {
				return err
			}

			if refs, ok := h.ChartRegistry.Get(d.GetID()); ok {
				refs.RebuildToolbar()
			}
			win.Close()
			return nil
		},
		func() {
			win.Close()
		},
	)

	win.SetContent(wrap(editor.Object()))
	win.Show()
}

func (h *DiagramActionHandler) validateExpression(expr string, d port.IDiagram) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	vars := d.CollectVariables()
	err := utils.AnalyzeExpressionWithRespectToTheDeclaredVariables(vars, expr)
	if err != nil {
		return err
	}

	return nil
}

func (h *DiagramActionHandler) applyFilterChanges(d port.IDiagram, changes *[]models.FilterComponentChange, currentMode models.FilterMode) {
	if x, ok := h.ChartRegistry.Get(d.GetID()); ok {
		if chart, ok := x.GetMainChart(); ok {
			theWidget, ok := chart.(diaw.DiagramWidget)
			if ok {
				theWidget.GetDataSeries().ApplyFilterChanges(*changes, currentMode)
			}
		}
		if c, ok := x.GetMainChart(); ok {
			c.Refresh()
		}
	}
}

func (h *DiagramActionHandler) saveFilters(jobID string, setupItem *models.SetupItem, filter *models.Filter, d port.IDiagram) error {
	if setupItem == nil {
		return fmt.Errorf("setup item is nil")
	}

	var newComponents []models.FilterComponent
	newMode := models.FilterModeAND
	if filter != nil && filter.Setup != nil {
		newComponents = filter.Setup.Components
		newMode = filter.Setup.Mode
	}

	changes := models.DiffFilterComponents(h.snapshotComponents, newComponents)

	if len(changes) > 0 || newMode != h.snaphshotMode {
		//handle expressions
		setupItem.HandleFilterChange(newComponents, changes, filter)
		//handle condition results
		h.applyFilterChanges(d, &changes, newMode)
	}

	owner := d.GetDiagramOwner()
	switch p := owner.(type) {
	case *vp.ViewProject:
		return p.Save()
	case *vp.DashboardProject:
		return p.Save()
	default:
		return fmt.Errorf("unsupported diagram owner")
	}
}
