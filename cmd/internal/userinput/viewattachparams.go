package userinput

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/gclkaze/evamon/pkg/utils"
)

type ViewAttachParams struct {
	JobID string

	// Path to the ViewWindow JSON file (--widget-path / -w)
	WidgetPath string

	// Filled after validation
	ViewWindow *viewproject.ViewWindow
}

func NewViewAttachParams(jobID, widgetPath string) (*ViewAttachParams, error) {
	p := &ViewAttachParams{
		JobID:      strings.TrimSpace(jobID),
		WidgetPath: strings.TrimSpace(widgetPath),
	}

	// Normalize widget path early
	if p.WidgetPath != "" {
		abs, err := filepath.Abs(p.WidgetPath)
		if err != nil {
			return nil, fmt.Errorf("abs(widget-path): %w", err)
		}
		p.WidgetPath = abs
	}

	return p, nil
}

func (p *ViewAttachParams) GetFileName() string {
	if p == nil || p.WidgetPath == "" {
		return ""
	}
	return filepath.Base(p.WidgetPath)
}

func (p *ViewAttachParams) IsValid() error {
	if p == nil {
		return fmt.Errorf("nil ViewAttachParams")
	}

	// --- job id ---
	if p.JobID == "" {
		return fmt.Errorf("--job-id is required")
	}

	// Prevent path traversal / separators
	if strings.Contains(p.JobID, "/") || strings.Contains(p.JobID, `\`) || strings.Contains(p.JobID, "..") {
		return fmt.Errorf("invalid --job-id")
	}

	// --- widget path ---
	if p.WidgetPath == "" {
		return fmt.Errorf("--widget-path is required")
	}

	// Use YOUR helper
	if !utils.FileExists(p.WidgetPath) {
		return fmt.Errorf("%s is not a file", p.WidgetPath)
	}

	// --- load + validate ViewWindow ---
	vw, err := viewproject.LoadViewWindow(p.WidgetPath)
	if err != nil {
		return err
	}

	p.ViewWindow = vw
	return nil
}

// Builds the final ViewProject after validation
func (p *ViewAttachParams) ToViewProject() (*viewproject.ViewProject, error) {
	if p.ViewWindow == nil {
		return nil, fmt.Errorf("view window not loaded (call IsValid first)")
	}

	vp := &viewproject.ViewProject{
		JobID: p.JobID,
		View:  *p.ViewWindow,
	}

	if err := vp.Validate(); err != nil {
		return nil, err
	}

	return vp, nil
}
