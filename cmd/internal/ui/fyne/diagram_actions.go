package fynerenderer

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/gclkaze/evamon/cmd/internal/models"
	diaw "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	"github.com/gclkaze/evamon/cmd/internal/ui/fyne/operations"

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

// maximizeViewProvider is implemented by BarChartWidget and LineChartWidget.
//
// MaximizeView(w, h) pre-sizes the drawer and installs a refreshHook so every
// drawer refresh call goes through adapter.Refresh() → CanvasForObject(adapter)
// (reliable — adapter IS window content) instead of CanvasForObject(d.root)
// which can be stale/nil when d.root lives inside a widget renderer.
//
// ClearMaximizeHook() removes the hook when the maximize window closes so that
// normal rendering (original canvas) resumes correctly.
type maximizeViewProvider interface {
	MaximizeView(initialW, initialH float32) fyne.CanvasObject
	ClearMaximizeHook()
}

func (h *DiagramActionHandler) Maximize(jobID string, d port.IDiagram) {
	refs, ok := h.ChartRegistry.Get(d.GetID())
	if !ok || refs.Maximized {
		return // no refs, or window already open
	}
	if refs.Chart == nil || refs.ChartSlot == nil {
		return
	}

	provider, ok := refs.Chart.(maximizeViewProvider)
	if !ok {
		return
	}

	chartObj := refs.Chart.Native().(fyne.CanvasObject)
	slotContainer := refs.ChartSlot.Native().(*fyne.Container)

	// Replace chart widget in the slot with a placeholder.
	slotContainer.RemoveAll()
	slotContainer.Add(container.NewCenter(
		widget.NewLabelWithStyle(
			"Maximized window is open",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true},
		),
	))

	const maxW, maxH float32 = 1200, 800

	// MaximizeView pre-sizes the drawer and installs the refresh hook.
	view := provider.MaximizeView(maxW, maxH)

	title := d.GetName()
	win := fyne.CurrentApp().NewWindow(title)
	win.Resize(fyne.NewSize(maxW, maxH))
	win.SetContent(view)

	refs.Maximized = true

	win.SetOnClosed(func() {
		provider.ClearMaximizeHook()
		slotContainer.RemoveAll()
		slotContainer.Add(chartObj)
		refs.Maximized = false
	})

	win.Show()
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

			/*			if refs, ok := h.ChartRegistry.Get(d.GetID()); ok {
						refs.RebuildToolbar()
						refs.RebuildFilterRow()
					}*/
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

func (h *DiagramActionHandler) Operations(jobID string, d port.IDiagram) {
	setups := d.GetSetup()
	if d == nil || len(setups) == 0 {
		return
	}
	setupItem := &setups[0]

	var components []models.FilterComponent
	if setupItem.Filter != nil && setupItem.Filter.Setup != nil {
		components = setupItem.Filter.Setup.Components
	}

	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) == 0 {
		return
	}
	parent := windows[0]

	operations.ShowOperationsModal(parent, components, []*models.TriggerRule{}, func(items []models.TriggerRule) {
		_ = items
	})
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

func (h *DiagramActionHandler) applyFilterChanges(d port.IDiagram, changes *[]models.FilterComponentChange) {
	if x, ok := h.ChartRegistry.Get(d.GetID()); ok {
		if chart, ok := x.GetMainChart(); ok {
			theWidget, ok := chart.(diaw.DiagramWidget)
			if ok {
				theWidget.GetDataSeries().ApplyFilterChanges(*changes)
			}
			chart.Refresh()
		}
	}
}

func (h *DiagramActionHandler) SetFilterComponentEnabled(jobID string, d port.IDiagram, componentID string, enabled bool) {
	setups := d.GetSetup()
	if d == nil || len(setups) == 0 {
		return
	}

	setupItem := &setups[0]
	if setupItem.Filter == nil || setupItem.Filter.Setup == nil {
		return
	}

	for i := range setupItem.Filter.Setup.Components {
		if setupItem.Filter.Setup.Components[i].ID == componentID {
			setupItem.Filter.Setup.Components[i].Enabled = enabled
			break
		}
	}

	h.RefreshDiagram(d)

	owner := d.GetDiagramOwner()
	switch p := owner.(type) {
	case *vp.ViewProject:
		p.Save()
	case *vp.DashboardProject:
		p.Save()
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
		setupItem.HandleFilterChange(changes, newMode)
		//handle condition results
		h.applyFilterChanges(d, &changes)
		if refs, ok := h.ChartRegistry.Get(d.GetID()); ok {
			refs.RebuildWrapper()
		}
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
