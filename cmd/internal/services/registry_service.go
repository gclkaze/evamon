package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/gclkaze/evamon/cmd/internal/fs"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
)

type ProjectsRegistryService struct {
	mu   sync.Mutex
	path string
	doc  projectsDoc
}

type projectsDoc struct {
	// id -> absolute file path of the stored project
	Projects  map[string]string `json:"projects"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

func NewProjectsRegistry(setup MainSetup) (*ProjectsRegistryService, error) {
	programDataLibDir := setup.GetWidgetPath()
	if programDataLibDir == "" {
		return nil, fmt.Errorf("lib dir is empty")
	}
	props := setup.GetProperties()
	fileName := props.GetString("projects_index_file_name", "projects.json")
	if strings.TrimSpace(fileName) == "" {
		fileName = "projects.json"
	}

	path := filepath.Join(programDataLibDir, fileName)

	r := &ProjectsRegistryService{
		path: path,
		doc: projectsDoc{
			Projects:  map[string]string{},
			UpdatedAt: time.Time{},
		},
	}

	// Load existing file if present
	if err := r.loadLocked(); err != nil {
		return nil, err
	}

	return r, nil
}

// Upsert adds or updates id -> filePath.
func (r *ProjectsRegistryService) Upsert(id string, filePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id == "" {
		return fmt.Errorf("id is empty")
	}
	if filePath == "" {
		return fmt.Errorf("filePath is empty")
	}

	abs, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("abs(filePath): %w", err)
	}

	// Ensure map exists
	if r.doc.Projects == nil {
		r.doc.Projects = map[string]string{}
	}

	r.doc.Projects[id] = abs
	r.doc.UpdatedAt = time.Now()

	return r.saveLocked()
}

// Delete removes an id. If id doesn't exist, it's NOT an error.
func (r *ProjectsRegistryService) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id == "" {
		return fmt.Errorf("id is empty")
	}

	if r.doc.Projects == nil {
		r.doc.Projects = map[string]string{}
	}

	delete(r.doc.Projects, id)
	r.doc.UpdatedAt = time.Now()

	return r.saveLocked()
}

// Get returns (path, true) if found.
func (r *ProjectsRegistryService) Get(id string) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id == "" {
		return "", false, fmt.Errorf("id is empty")
	}
	if r.doc.Projects == nil {
		return "", false, nil
	}
	p, ok := r.doc.Projects[id]
	return p, ok, nil
}

// List returns a copy of the mapping.
func (r *ProjectsRegistryService) List() (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make(map[string]string, len(r.doc.Projects))
	for k, v := range r.doc.Projects {
		out[k] = v
	}
	return out, nil
}

// Reload forces a re-read from disk (rarely needed).
func (r *ProjectsRegistryService) Reload() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadLocked()
}

// ----------------- internal -----------------

func (r *ProjectsRegistryService) loadLocked() error {
	b, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			// nothing to load; start empty
			return nil
		}
		return fmt.Errorf("read %q: %w", r.path, err)
	}

	var doc projectsDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return fmt.Errorf("parse %q: %w", r.path, err)
	}

	if doc.Projects == nil {
		doc.Projects = map[string]string{}
	}

	r.doc = doc
	return nil
}

func (r *ProjectsRegistryService) saveLocked() error {
	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(r.path), err)
	}

	b, err := json.MarshalIndent(r.doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal projects registry: %w", err)
	}

	// Your atomic writer
	if err := fs.WriteFileAtomic(r.path, b, 0o644); err != nil {
		return err
	}
	return nil
}

// Print prints the project registry in a "docker ps style" table by default,
// or in a Go-template format if params.Format is set.
//
// Supported filters (-f/--filter key=value, repeatable):
//   - id / project / projectid: substring match on project id
//   - job / jobid:             substring match on job id
//   - status:                  OK | MISSING | INVALID (case-insensitive)
//   - type / types:            substring match on types summary (e.g. boolean, bar)
//   - multitab:                true/false/1/0/yes/no
//   - file / base:             substring match on base filename
//   - path:                    substring match on full path
//
// Row fields for --format template:
//
//	.ProjectID .JobID .MultiTab .Diagrams .Types .File .Path .Status .Base
func (inst *ProjectsRegistryService) Print(params *userinput.ViewLsParams, logger output.Printer) error {
	if inst == nil {
		return fmt.Errorf("nil ProjectsRegistryService")
	}
	if params == nil {
		return fmt.Errorf("nil params")
	}
	if logger == nil {
		return fmt.Errorf("nil logger")
	}
	if err := params.IsValid(); err != nil {
		return err
	}

	// Get snapshot of registry map (id -> file path)
	entries := inst.doc.Projects
	if entries == nil {
		entries = map[string]string{}
	}
	if len(entries) == 0 {
		logger.Info("No view projects found.")
		return nil
	}

	fset, err := parseViewLsFilters(params.Filters)
	if err != nil {
		return err
	}

	rows := buildRowsFromRegistry(entries)

	// Apply filters
	rows = applyViewLsFilters(rows, fset)

	// Hide INVALID/MISSING by default unless --all
	if !params.ShowAll {
		rows = filterByStatus(rows, map[string]bool{"OK": true})
	}

	if len(rows) == 0 {
		logger.Info("No matching view projects.")
		return nil
	}

	// Output
	if strings.TrimSpace(params.Format) != "" {
		out, err := renderFormat(rows, params.Format)
		if err != nil {
			return err
		}
		logger.Info(out)
		return nil
	}

	out := renderTable(rows, params.NoTrunc)
	logger.Info(out)
	return nil
}

// -------------------- row model --------------------

type viewLsRow struct {
	ProjectID string
	JobID     string
	MultiTab  bool
	Diagrams  int
	Types     string
	File      string
	Path      string
	Status    string // OK | MISSING | INVALID
	Base      string
}

// -------------------- building rows --------------------

func buildRowsFromRegistry(reg map[string]string) []viewLsRow {
	ids := make([]string, 0, len(reg))
	for id := range reg {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	rows := make([]viewLsRow, 0, len(ids))
	for _, id := range ids {
		p := reg[id]
		base := filepath.Base(p)

		r := viewLsRow{
			ProjectID: id,
			JobID:     "-",
			MultiTab:  false,
			Diagrams:  0,
			Types:     "-",
			File:      base,
			Path:      p,
			Base:      base,
			Status:    "OK",
		}

		if _, statErr := os.Stat(p); statErr != nil {
			r.Status = "MISSING"
			rows = append(rows, r)
			continue
		}

		vp, loadErr := viewproject.LoadViewProject(p)
		if loadErr != nil {
			r.Status = "INVALID"
			rows = append(rows, r)
			continue
		}

		r.JobID = vp.JobID
		r.MultiTab = vp.View.MultiTab
		r.Diagrams = len(vp.View.Diagrams)
		r.Types = summarizeDiagramTypes(vp)
		vp.BindDiagramPointers()
		rows = append(rows, r)
	}

	return rows
}

func summarizeDiagramTypes(vp *viewproject.ViewProject) string {
	counts := map[string]int{}
	for _, d := range vp.View.Diagrams {
		t := strings.TrimSpace(string(d.Type))
		if t == "" {
			t = "unknown"
		}
		counts[t]++
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s(%d)", k, counts[k]))
	}
	return strings.Join(parts, ",")
}

// -------------------- filters --------------------

type filterSet map[string][]string

func parseViewLsFilters(filters []string) (filterSet, error) {
	fs := filterSet{}
	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid filter %q (expected key=value)", f)
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			return nil, fmt.Errorf("invalid filter %q (empty key or value)", f)
		}
		fs[k] = append(fs[k], v)
	}
	return fs, nil
}

func applyViewLsFilters(rows []viewLsRow, fs filterSet) []viewLsRow {
	if len(fs) == 0 {
		return rows
	}
	out := make([]viewLsRow, 0, len(rows))
	for _, r := range rows {
		if matchesAllFilters(r, fs) {
			out = append(out, r)
		}
	}
	return out
}

func matchesAllFilters(r viewLsRow, fs filterSet) bool {
	for k, values := range fs {
		if !matchesAnyValue(r, k, values) {
			return false
		}
	}
	return true
}

func matchesAnyValue(r viewLsRow, key string, values []string) bool {
	switch key {
	case "id", "project", "projectid":
		return anyMatch(values, func(v string) bool { return strings.Contains(r.ProjectID, v) })

	case "job", "jobid":
		return anyMatch(values, func(v string) bool { return strings.Contains(r.JobID, v) })

	case "status":
		return anyMatch(values, func(v string) bool { return strings.EqualFold(r.Status, v) })

	case "type", "types":
		return anyMatch(values, func(v string) bool {
			return strings.Contains(strings.ToLower(r.Types), strings.ToLower(v))
		})

	case "multitab":
		return anyMatch(values, func(v string) bool {
			b, ok := parseBoolLoose(v)
			return ok && r.MultiTab == b
		})

	case "file", "base":
		return anyMatch(values, func(v string) bool { return strings.Contains(r.Base, v) })

	case "path":
		return anyMatch(values, func(v string) bool { return strings.Contains(r.Path, v) })

	default:
		// Unknown filter key => strict mismatch
		return false
	}
}

func anyMatch(values []string, fn func(v string) bool) bool {
	for _, v := range values {
		if fn(v) {
			return true
		}
	}
	return false
}

func parseBoolLoose(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "y":
		return true, true
	case "false", "0", "no", "n":
		return false, true
	default:
		return false, false
	}
}

func filterByStatus(rows []viewLsRow, allowed map[string]bool) []viewLsRow {
	out := make([]viewLsRow, 0, len(rows))
	for _, r := range rows {
		if allowed[r.Status] {
			out = append(out, r)
		}
	}
	return out
}

// -------------------- rendering --------------------

func renderTable(rows []viewLsRow, noTrunc bool) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "PROJECT ID\tJOB ID\tMULTITAB\tDIAGRAMS\tTYPES\tFILE\tSTATUS")

	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%v\t%d\t%s\t%s\t%s\n",
			r.ProjectID, // ← NEVER truncate
			trunc(r.JobID, 18, noTrunc),
			r.MultiTab,
			r.Diagrams,
			trunc(r.Types, 22, noTrunc),
			trunc(r.Base, 28, noTrunc),
			r.Status,
		)
	}

	_ = w.Flush()
	return strings.TrimRight(buf.String(), "\n")
}

func renderFormat(rows []viewLsRow, format string) (string, error) {
	tpl, err := template.New("format").Parse(format)
	if err != nil {
		return "", fmt.Errorf("invalid --format template: %w", err)
	}

	needsNL := !strings.HasSuffix(format, "\n")

	var buf bytes.Buffer
	for _, r := range rows {
		if err := tpl.Execute(&buf, r); err != nil {
			return "", fmt.Errorf("execute template: %w", err)
		}
		if needsNL {
			buf.WriteByte('\n')
		}
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

func trunc(s string, max int, noTrunc bool) string {
	if noTrunc || max <= 0 {
		return s
	}
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	if max <= 1 {
		return string(rs[:max])
	}
	return string(rs[:max-1]) + "…"
}
