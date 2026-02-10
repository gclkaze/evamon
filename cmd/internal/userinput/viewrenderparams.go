package userinput

import (
	"fmt"
	"strings"
)

type ViewRenderParams struct {
	ProjectID string
	Headless  bool
}

func NewViewRenderParams(projectID string, headless bool) *ViewRenderParams {
	return &ViewRenderParams{
		ProjectID: strings.TrimSpace(projectID),
		Headless:  headless,
	}
}

func (p *ViewRenderParams) IsValid() error {
	if p == nil {
		return fmt.Errorf("nil ViewRenderParams")
	}
	if p.ProjectID == "" {
		return fmt.Errorf("--project-id is required")
	}
	return nil
}
