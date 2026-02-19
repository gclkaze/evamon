package userinput

import (
	"fmt"
	"strings"

	"github.com/gclkaze/evamon/pkg/utils"
)

type ViewDashboardRenderParams struct {
	Path string
}

func NewViewDashboardRenderParams(path string) *ViewDashboardRenderParams {
	return &ViewDashboardRenderParams{
		Path: strings.TrimSpace(path),
	}
}

func (p *ViewDashboardRenderParams) IsValid() error {

	if p.Path == "" {
		return fmt.Errorf("the Dashboard project path parameter required")
	}

	if !utils.FileExists(p.Path) {
		return fmt.Errorf("the provided Dashboard project path parameter does not exist '%s'", p.Path)
	}
	return nil
}
