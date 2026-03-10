package fynerenderer

import (
	"fmt"
	"log"
	"strings"

	vp "github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/gclkaze/evamon/pkg/utils"
)

type DiagramActionHandler struct {
	// add dependencies later:
	// WindowManager
	// ExportService
	// Filter validator / parser
	// Project saver
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

func (h DiagramActionHandler) SetFilterEnabled(jobID string, d *vp.Diagram, enabled bool) {
	if d == nil || len(d.Setup) == 0 {
		return
	}

	setupItem := &d.Setup[0]
	if setupItem.Filter == nil {
		return
	}

	setupItem.Filter.Enabled = enabled

	// TODO: persist project/config if needed
	// TODO: trigger chart refresh/re-filter if needed

	log.Printf(
		"set filter enabled=%v for job=%s diagram=%s\n",
		enabled, jobID, d.GetName(),
	)
}

func (h DiagramActionHandler) Filters(jobID string, d *vp.Diagram) {
	log.Printf("filters clicked for job=%s diagram=%s\n", jobID, d.GetName())

	if d == nil || len(d.Setup) == 0 {
		return
	}

	win := NewChildWindow("Filters", 520, 360)
	setupItem := &d.Setup[0]

	editor := NewFilterEditor(
		"Filters - "+d.GetName(),
		setupItem.Filter,
		d,
		func(expr string) error {
			return h.validateExpression(expr, d)
		},
		func(f *vp.Filter, d *vp.Diagram) error {
			if err := h.saveFilters(jobID, setupItem, f, d); err != nil {
				return err
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

func (h DiagramActionHandler) validateExpression(expr string, d *vp.Diagram) error {
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

func (h DiagramActionHandler) saveFilters(jobID string, setupItem *vp.SetupItem, filter *vp.Filter, d *vp.Diagram) error {
	if setupItem == nil {
		return fmt.Errorf("setup item is nil")
	}

	setupItem.Filter = filter

	owner := d.Owner
	switch p := owner.(type) {
	case *vp.ViewProject:
		return p.Save()
	case *vp.DashboardProject:
		return p.Save()
	default:
		return fmt.Errorf("unsupported diagram owner")
	}

	return nil
}
