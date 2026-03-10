package models

import (
	"fmt"

	"github.com/gclkaze/evamon/pkg/utils"
)

type ViewWindow struct {
	MultiTab    bool         `json:"multiTab"`
	WindowStyle *WindowStyle `json:"windowStyle,omitempty"`
	Diagrams    []Diagram    `json:"diagrams"`
}

// Validate validates the view-only document.
func (vw ViewWindow) Validate() error {
	if len(vw.Diagrams) == 0 {
		return fmt.Errorf("diagrams must not be empty")
	}
	if err := ValidateWindowStyle("windowStyle", vw.WindowStyle); err != nil {
		return err
	}

	for i := range vw.Diagrams {
		if err := vw.Diagrams[i].Validate(i); err != nil {
			return fmt.Errorf("diagrams[%d]: %w", i, err)
		}
	}
	return nil
}
func EnsureDiagramIDs(diagrams []Diagram) bool {
	modified := false
	seen := make(map[string]struct{})

	for i := range diagrams {
		id := diagrams[i].GetID()

		if id == "" {
			id = newUniqueDiagramID(seen)
			diagrams[i].SetID(id)
			modified = true
		} else if _, exists := seen[id]; exists {
			id = newUniqueDiagramID(seen)
			diagrams[i].SetID(id)
			modified = true
		}

		seen[id] = struct{}{}
	}

	return modified
}
func newUniqueDiagramID(seen map[string]struct{}) string {
	for {
		id := utils.GetRandomString()
		if _, exists := seen[id]; !exists {
			return id
		}
	}
}

func ValidateWindowStyle(path string, ws *WindowStyle) error {
	if ws == nil {
		return nil
	}
	if ws.Width != nil && *ws.Width <= 0 {
		return fmt.Errorf("%s.width must be > 0", path)
	}
	if ws.Height != nil && *ws.Height <= 0 {
		return fmt.Errorf("%s.height must be > 0", path)
	}
	return nil
}
