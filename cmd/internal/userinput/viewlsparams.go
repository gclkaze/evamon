package userinput

import (
	"fmt"
	"strings"
)

type ViewLsParams struct {
	ShowAll bool     // --all / -a
	NoTrunc bool     // --no-trunc
	Format  string   // --format
	Filters []string // --filter / -f (repeatable)
}

func NewViewLsParams(showAll, noTrunc bool, format string, filters []string) *ViewLsParams {
	// Normalize a little (optional)
	f := make([]string, 0, len(filters))
	for _, x := range filters {
		x = strings.TrimSpace(x)
		if x != "" {
			f = append(f, x)
		}
	}

	return &ViewLsParams{
		ShowAll: showAll,
		NoTrunc: noTrunc,
		Format:  strings.TrimSpace(format),
		Filters: f,
	}
}

func (p *ViewLsParams) IsValid() error {
	if p == nil {
		return fmt.Errorf("nil ViewLsParams")
	}

	// You can keep this permissive, but at least validate filter shape key=value.
	for _, f := range p.Filters {
		if !strings.Contains(f, "=") {
			return fmt.Errorf("invalid filter %q (expected key=value)", f)
		}
		k, v, ok := strings.Cut(f, "=")
		if !ok || strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			return fmt.Errorf("invalid filter %q (expected key=value)", f)
		}
	}

	// format can be validated by the Print() method when it parses the template
	return nil
}
