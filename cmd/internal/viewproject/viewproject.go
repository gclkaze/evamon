package viewproject

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gclkaze/evamon/cmd/internal/fs"
	models "github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/pkg/utils"
)

// ============================================================
// Models: ViewWindow (view-only) + ViewProject (jobID + view)
// ============================================================

// ViewWindow is the reusable view-only document (no jobID).
// It corresponds to JSON like:
//
//	{
//	  "multiTab": false,
//	  "diagrams": [ ... ]
//	}

// ViewProject wraps a ViewWindow and targets a JobID.
// It corresponds to JSON like:
//
//	{
//	  "jobID": "this-is-an-id",
//	  "view": { ...ViewWindow... }
//	}
type ViewProject struct {
	models.ProjectBase

	JobID string            `json:"jobID"`
	View  models.ViewWindow `json:"view"`
}

func (vp *ViewProject) GetProjectBase() *models.ProjectBase {
	return &vp.ProjectBase
}

func (vp *ViewProject) GetOwnerKind() string {
	return "viewProject"
}

func (vp *ViewProject) BindDiagramPointers() {
	for i := range vp.View.Diagrams {
		vp.View.Diagrams[i].Owner = vp
	}
}

// --------------------
// Typed union for diagramStyle
// --------------------
func (inst ViewProject) CollectVariables() []string {
	set := make(map[string]struct{})

	for _, d := range inst.View.Diagrams {
		for _, v := range d.CollectVariables() {
			set[v] = struct{}{}
		}
	}

	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}

	return out
}

// ============================================================
// Loaders
// ============================================================

// LoadViewWindow reads a JSON ViewWindow file from disk, parses it strictly,
// and validates the result.
func LoadViewWindow(path string) (*models.ViewWindow, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("view path is empty")
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

	var vw models.ViewWindow
	if err := utils.DecodeStrict(data, &vw); err != nil {
		return nil, fmt.Errorf("parse %q: %w", absPath, err)
	}

	if err := vw.Validate(); err != nil {
		return nil, fmt.Errorf("invalid view %q: %w", absPath, err)
	}

	return &vw, nil
}

// LoadViewProject reads a JSON ViewProject file from disk, parses it strictly,
// and validates the result. It also fills vp.ProjectPath.
func LoadViewProject(path string) (*ViewProject, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("project path is empty")
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

	var vp ViewProject
	if err := utils.DecodeStrict(data, &vp); err != nil {
		return nil, fmt.Errorf("parse %q: %w", absPath, err)
	}

	vp.ProjectPath = absPath

	modified := vp.EnsureDiagramIDs()
	if modified {
		if err := fs.SaveProjectJSON(vp.ProjectPath, &vp); err != nil {
			return nil, fmt.Errorf("save normalized dashboard %q: %w", absPath, err)
		}
	}

	if err := vp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project %q: %w", absPath, err)
	}

	return &vp, nil
}
func newUniqueDiagramID(seen map[string]struct{}) string {
	for {
		id := utils.GetRandomString()
		if _, exists := seen[id]; !exists {
			return id
		}
	}
}
func (vp *ViewProject) EnsureDiagramIDs() bool {
	if vp == nil {
		return false
	}

	modified := false
	seen := make(map[string]struct{})

	for i := range vp.View.Diagrams {
		id := strings.TrimSpace(vp.View.Diagrams[i].GetID())

		if id == "" {
			id = newUniqueDiagramID(seen)
			vp.View.Diagrams[i].SetID(id)
			modified = true
		} else if _, exists := seen[id]; exists {
			id = newUniqueDiagramID(seen)
			vp.View.Diagrams[i].SetID(id)
			modified = true
		}

		seen[id] = struct{}{}
	}

	return modified
}

// expandUser supports "~" and "~/" on Unix-like environments.
// On Windows, "~" isn't typically used but the function is harmless.
func expandUser(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

func (vp *ViewProject) Save() error {
	if vp == nil {
		return fmt.Errorf("view project is nil")
	}
	if err := vp.Validate(); err != nil {
		return fmt.Errorf("validate before save: %w", err)
	}
	return fs.SaveProjectJSON(vp.ProjectPath, vp)
}

// ============================================================
// Custom unmarshalling for typed diagramStyle union
// ============================================================

// Internal helper type for decoding setup items with raw diagramStyle.

// ============================================================
// Validation
// ============================================================

// Validate validates the project wrapper (jobID + view).
func (vp ViewProject) Validate() error {
	if strings.TrimSpace(vp.JobID) == "" {
		return fmt.Errorf("jobID is required")
	}
	if err := vp.View.Validate(); err != nil {
		return fmt.Errorf("view: %w", err)
	}
	return nil
}
