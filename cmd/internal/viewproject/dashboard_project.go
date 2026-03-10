package viewproject

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gclkaze/evamon/cmd/internal/fs"
	models "github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/pkg/utils"
)

// ============================================================
// Models: DashboardProject (ID + title + views grid)
// Reuses ViewWindow, WindowStyle, Diagram
// ============================================================

type DashboardProject struct {
	models.ProjectBase

	Title       string              `json:"title"`
	WindowStyle *models.WindowStyle `json:"windowStyle,omitempty"`
	Views       []DashboardView     `json:"views"`
}

func (dp *DashboardProject) GetProjectBase() *models.ProjectBase {
	return &dp.ProjectBase
}

func (dp *DashboardProject) GetOwnerKind() string {
	return "dashboardProject"
}

type DashboardView struct {
	ID          string              `json:"id,omitempty"`
	Title       string              `json:"title"`
	WindowStyle *models.WindowStyle `json:"windowStyle,omitempty"`
	Rows        []DashboardRow      `json:"rows"`
}

type DashboardRow struct {
	Columns []DashboardColumn `json:"columns"`
}

func (vp *DashboardProject) BindProjectPointers() {
	for i := range vp.Views {
		for r := range vp.Views[i].Rows {
			for c := range vp.Views[i].Rows[r].Columns {
				viewWindow := vp.Views[i].Rows[r].Columns[c].View
				for d := range viewWindow.Diagrams {
					vp.Views[i].Rows[r].Columns[c].View.Diagrams[d].Owner = vp
				}
			}
		}
	}
}

type DashboardColumn struct {
	// Optional future extension (e.g., grid span). Safe to ignore for now.
	Width *int `json:"width,omitempty"`

	JobID string `json:"jobId"`

	// Accepts either:
	// - "widget": { "type": "...", "setup": [...] }   (your generated JSON)
	// - "view":   { "multiTab": false, "diagrams": [...] } (native reuse form)
	//
	// Internally we normalize into ViewWindow so we reuse existing Diagram parsing/validation.
	View models.ViewWindow `json:"view"`
}

// Raw column decoding helper for accepting both "widget" and "view".
type dashboardColumnRaw struct {
	Width  *int               `json:"width,omitempty"`
	JobID  string             `json:"jobId"`
	Widget *models.Diagram    `json:"widget,omitempty"`
	View   *models.ViewWindow `json:"view,omitempty"`
}

// Custom unmarshalling: if "widget" is present, wrap it as ViewWindow{MultiTab:false, Diagrams:[widget]}
func (c *DashboardColumn) UnmarshalJSON(b []byte) error {
	var raw dashboardColumnRaw
	if err := utils.DecodeStrict(b, &raw); err != nil {
		return err
	}
	if strings.TrimSpace(raw.JobID) == "" {
		return fmt.Errorf("jobId is required")
	}
	c.Width = raw.Width
	c.JobID = raw.JobID

	switch {
	case raw.View != nil && raw.Widget != nil:
		return fmt.Errorf("provide either 'view' or 'widget', not both")
	case raw.View != nil:
		c.View = *raw.View
		return nil
	case raw.Widget != nil:
		c.View = models.ViewWindow{
			MultiTab: false,
			Diagrams: []models.Diagram{*raw.Widget},
		}
		return nil
	default:
		return fmt.Errorf("either 'view' or 'widget' is required")
	}
}

// ============================================================
// Loaders
// ============================================================

func LoadDashboardProject(path string) (*DashboardProject, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("dashboard project path is empty")
	}

	expanded, err := expandUser(path)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return nil, fmt.Errorf("abs(%q): %w", expanded, err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", absPath, err)
	}

	var dp DashboardProject
	if err := utils.DecodeStrict(data, &dp); err != nil {
		return nil, fmt.Errorf("parse %q: %w", absPath, err)
	}

	dp.ProjectPath = absPath

	if err := dp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dashboard %q: %w", absPath, err)
	}
	modified := dp.EnsureDiagramIDs()
	if modified {
		if err := fs.SaveProjectJSON(dp.ProjectPath, &dp); err != nil {
			return nil, fmt.Errorf("save normalized dashboard %q: %w", absPath, err)
		}
	}

	return &dp, nil
}
func (dp *DashboardProject) EnsureDiagramIDs() bool {
	if dp == nil {
		return false
	}

	modified := false
	seen := make(map[string]struct{})

	for vi := range dp.Views {
		for ri := range dp.Views[vi].Rows {
			for ci := range dp.Views[vi].Rows[ri].Columns {
				view := &dp.Views[vi].Rows[ri].Columns[ci].View
				if models.EnsureDiagramIDsInViewWindow(view, seen) {
					modified = true
				}
			}
		}
	}

	return modified
}
func (dp *DashboardProject) Save() error {
	if dp == nil {
		return fmt.Errorf("dashboard project is nil")
	}
	if err := dp.Validate(); err != nil {
		return fmt.Errorf("validate before save: %w", err)
	}
	return fs.SaveProjectJSON(dp.ProjectPath, dp)
}

// ============================================================
// Validation
// ============================================================

func (dp *DashboardProject) Validate() error {
	if strings.TrimSpace(dp.ID) == "" {
		return fmt.Errorf("ID is required")
	}
	if strings.TrimSpace(dp.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if err := models.ValidateWindowStyle("windowStyle", dp.WindowStyle); err != nil {
		return err
	}
	if len(dp.Views) == 0 {
		return fmt.Errorf("views must not be empty")
	}

	for i := range dp.Views {
		if err := dp.Views[i].Validate(i); err != nil {
			return fmt.Errorf("views[%d]: %w", i, err)
		}
	}
	return nil
}

func (v DashboardView) Validate(viewIndex int) error {
	if strings.TrimSpace(v.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if err := models.ValidateWindowStyle("windowStyle", v.WindowStyle); err != nil {
		return err
	}
	if len(v.Rows) == 0 {
		return fmt.Errorf("rows must not be empty")
	}
	for i := range v.Rows {
		if err := v.Rows[i].Validate(i); err != nil {
			return fmt.Errorf("rows[%d]: %w", i, err)
		}
	}
	return nil
}

func (r DashboardRow) Validate(rowIndex int) error {
	if len(r.Columns) == 0 {
		return fmt.Errorf("columns must not be empty")
	}
	for j := range r.Columns {
		if err := r.Columns[j].Validate(j); err != nil {
			return fmt.Errorf("columns[%d]: %w", j, err)
		}
	}
	return nil
}

func (c DashboardColumn) Validate(colIndex int) error {
	if strings.TrimSpace(c.JobID) == "" {
		return fmt.Errorf("jobId is required")
	}
	if c.Width != nil && *c.Width <= 0 {
		return fmt.Errorf("width must be > 0")
	}
	// Reuse existing ViewWindow validation (and through it Diagram validation)
	if err := c.View.Validate(); err != nil {
		return fmt.Errorf("view: %w", err)
	}
	// For a single "widget" column, it’s common to require exactly 1 diagram,
	// but we won’t enforce it—your ViewWindow supports multiple diagrams.
	return nil
}

// GetUniqueJobIDs returns all distinct, non-empty JobIDs referenced
// in the dashboard, in stable sorted order.
func (dp *DashboardProject) GetUniqueJobIDs() ([]string, error) {
	if dp == nil {
		return nil, fmt.Errorf("dashboard is nil")
	}

	set := make(map[string]struct{})

	for vi := range dp.Views {
		v := dp.Views[vi]

		for ri := range v.Rows {
			r := v.Rows[ri]

			for ci := range r.Columns {
				c := r.Columns[ci]

				jobID := strings.TrimSpace(c.JobID)
				if jobID == "" {
					// If your validation already guarantees this,
					// you could skip instead of error.
					return nil, fmt.Errorf(
						"views[%d].rows[%d].columns[%d]: jobID is empty",
						vi, ri, ci,
					)
				}

				set[jobID] = struct{}{}
			}
		}
	}

	if len(set) == 0 {
		return nil, fmt.Errorf("no jobIDs found in dashboard")
	}

	// Convert to slice
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}

	// Stable order is important for:
	// - deterministic subscriptions
	// - tests
	// - logs
	sort.Strings(out)

	return out, nil
}

func (inst *DashboardProject) IsEmpty() bool {
	if len(inst.Views) == 0 {
		return true
	}

	for i := range inst.Views {
		view := inst.Views[i]
		if len(view.Rows) == 0 {
			continue
		}

		rows := view.Rows
		for j := range rows {
			row := rows[j]

			if len(row.Columns) == 0 {
				continue
			}

			return false
		}
	}

	return true
}
