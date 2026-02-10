// ui/factory.go
package ui

import (
	"fmt"

	fynerenderer "github.com/gclkaze/evamon/cmd/internal/ui/fyne"
	"github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Kind string

const (
	KindFyne Kind = "fyne"
	KindWeb  Kind = "web"
)

func NewRenderer(kind Kind) (port.Renderer, error) {
	switch kind {
	case KindFyne:
		return fynerenderer.New(), nil
	default:
		return nil, fmt.Errorf("unknown renderer: %s", kind)
	}
}
