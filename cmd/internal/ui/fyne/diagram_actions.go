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
		func(expr string) error {
			return h.validateExpression(expr)
		},
		func(f *vp.Filter) error {
			if err := h.saveFilters(jobID, setupItem, f); err != nil {
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

func (h DiagramActionHandler) validateExpression(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// TODO: call your DSL/filter parser here
	// If invalid, return parser error.

	res := utils.AnalyzeExpression(expr)
	if res {
		fmt.Print("OK")
	}

	return nil
}

func (h DiagramActionHandler) saveFilters(jobID string, setupItem *vp.SetupItem, filter *vp.Filter) error {
	if setupItem == nil {
		return fmt.Errorf("setup item is nil")
	}

	setupItem.Filter = filter

	// TODO: persist project/config here if needed.
	// Example:
	// return h.projectSaver.Save(...)

	return nil
}
