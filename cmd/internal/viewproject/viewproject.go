package viewproject

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
type ViewWindow struct {
	MultiTab    bool         `json:"multiTab"`
	WindowStyle *WindowStyle `json:"windowStyle,omitempty"`
	Diagrams    []Diagram    `json:"diagrams"`
}

// ViewProject wraps a ViewWindow and targets a JobID.
// It corresponds to JSON like:
//
//	{
//	  "jobID": "this-is-an-id",
//	  "view": { ...ViewWindow... }
//	}
type ViewProject struct {
	ProjectBase

	JobID string     `json:"jobID"`
	View  ViewWindow `json:"view"`
}

type WindowStyle struct {
	Width  *float32 `json:"width,omitempty"`
	Height *float32 `json:"height,omitempty"`
}

// Diagram contains a typed union for setup items based on Diagram.Type.
type Diagram struct {
	Type  DiagramType `json:"type"`
	Setup []SetupItem `json:"setup"`
}

type DiagramType string

const (
	DiagramTypeBoolean DiagramType = "boolean"
	DiagramTypeBar     DiagramType = "bar"
	DiagramTypeLine    DiagramType = "line"
)

// SetupItem is one variable/series to visualize.
type SetupItem struct {
	Variable     string       `json:"variable"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	VariableType ValueType    `json:"variableType"`
	WindowStyle  *WindowStyle `json:"windowStyle,omitempty"`

	// Typed union:
	// - for boolean diagrams: BooleanStyle
	// - for bar diagrams:     BarStyle
	DiagramStyle Style `json:"diagramStyle,omitempty"`

	MultiVariableSetup []MultiSetupItem `json:"multiVariableSetup,omitempty"`
}

type MultiSetupItem struct {
	Variable     string    `json:"variable"` // key inside bundle (cpu/mem/gpu)
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	VariableType ValueType `json:"variableType"`

	DiagramStyle Style `json:"diagramStyle,omitempty"`
}

type ValueType string

const (
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeInteger ValueType = "integer"
	ValueTypeFloat   ValueType = "float"
)

// --------------------
// Typed union for diagramStyle
// --------------------

type Style interface {
	isStyle()
}

// BooleanStyle matches JSON keys "true"/"false".
type BooleanStyle struct {
	True  string `json:"true"`
	False string `json:"false"`
}

func (BooleanStyle) isStyle() {}

// BarStyle matches JSON keys "axis"/"background".
type BarStyle struct {
	Axis       string `json:"axis,omitempty"`
	Background string `json:"background,omitempty"`

	Bar string `json:"bar,omitempty"`
}

func (BarStyle) isStyle() {}

type LineStyle struct {
	Line       string `json:"line,omitempty"`
	Background string `json:"background,omitempty"`
}

func (LineStyle) isStyle() {}

// ============================================================
// Loaders
// ============================================================

// LoadViewWindow reads a JSON ViewWindow file from disk, parses it strictly,
// and validates the result.
func LoadViewWindow(path string) (*ViewWindow, error) {
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

	var vw ViewWindow
	if err := decodeStrict(data, &vw); err != nil {
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
	if err := decodeStrict(data, &vp); err != nil {
		return nil, fmt.Errorf("parse %q: %w", absPath, err)
	}

	vp.ProjectPath = absPath

	if err := vp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project %q: %w", absPath, err)
	}

	return &vp, nil
}

// ============================================================
// Strict JSON helpers
// ============================================================

// decodeStrict decodes JSON and fails on unknown fields.
func decodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	dec.UseNumber()

	if err := dec.Decode(v); err != nil {
		return err
	}
	// Ensure no trailing junk
	if dec.More() {
		return errors.New("unexpected trailing JSON content")
	}
	return nil
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

// ============================================================
// Custom unmarshalling for typed diagramStyle union
// ============================================================

func (d *Diagram) UnmarshalJSON(b []byte) error {
	// Step 1: decode diagram envelope (type + setup raw)
	var raw struct {
		Type  DiagramType       `json:"type"`
		Setup []json.RawMessage `json:"setup"`
	}
	if err := decodeStrict(b, &raw); err != nil {
		return err
	}

	if raw.Type == "" {
		return fmt.Errorf("diagram.type is required")
	}

	d.Type = raw.Type
	d.Setup = make([]SetupItem, 0, len(raw.Setup))

	for i, itemRaw := range raw.Setup {
		var item setupItemRaw
		if err := decodeStrict(itemRaw, &item); err != nil {
			return fmt.Errorf("diagram.setup[%d]: %w", i, err)
		}

		out := SetupItem{
			Variable:     item.Variable,
			Title:        item.Title,
			Description:  item.Description,
			VariableType: item.VariableType,
			WindowStyle:  item.WindowStyle,
		}

		// Parse top-level diagramStyle (for single-variable diagrams OR “bundle-level” style)
		if len(item.DiagramStyle) > 0 && string(item.DiagramStyle) != "null" {
			style, err := parseStyleForDiagramType(raw.Type, item.DiagramStyle)
			if err != nil {
				return fmt.Errorf("diagram.setup[%d].diagramStyle: %w", i, err)
			}
			out.DiagramStyle = style
		}

		// NEW: parse multiVariableSetup entries (sub-series)
		if len(item.MultiVariableSetup) > 0 {
			out.MultiVariableSetup = make([]MultiSetupItem, 0, len(item.MultiVariableSetup))

			for k, subRaw := range item.MultiVariableSetup {
				var sub multiSetupItemRaw
				if err := decodeStrict(subRaw, &sub); err != nil {
					return fmt.Errorf("diagram.setup[%d].multiVariableSetup[%d]: %w", i, k, err)
				}

				subOut := MultiSetupItem{
					Variable:     sub.Variable,
					Title:        sub.Title,
					Description:  sub.Description,
					VariableType: sub.VariableType,
				}

				if len(sub.DiagramStyle) > 0 && string(sub.DiagramStyle) != "null" {
					subStyle, err := parseStyleForDiagramType(raw.Type, sub.DiagramStyle)
					if err != nil {
						return fmt.Errorf("diagram.setup[%d].multiVariableSetup[%d].diagramStyle: %w", i, k, err)
					}
					subOut.DiagramStyle = subStyle
				}

				out.MultiVariableSetup = append(out.MultiVariableSetup, subOut)
			}
		}

		d.Setup = append(d.Setup, out)
	}

	return nil
}

// Internal helper type for decoding setup items with raw diagramStyle.
type setupItemRaw struct {
	Variable    string `json:"variable"`
	Title       string `json:"title"`
	Description string `json:"description"`

	VariableType ValueType       `json:"variableType"`
	DiagramStyle json.RawMessage `json:"diagramStyle,omitempty"`
	WindowStyle  *WindowStyle    `json:"windowStyle,omitempty"`

	MultiVariableSetup []json.RawMessage `json:"multiVariableSetup,omitempty"`
}

type multiSetupItemRaw struct {
	Variable    string `json:"variable"`
	Title       string `json:"title"`
	Description string `json:"description"`

	VariableType ValueType       `json:"variableType"`
	DiagramStyle json.RawMessage `json:"diagramStyle,omitempty"`
}

func parseStyleForDiagramType(t DiagramType, raw json.RawMessage) (Style, error) {
	switch t {
	case DiagramTypeBoolean:
		var s BooleanStyle
		if err := decodeStrict(raw, &s); err != nil {
			return nil, err
		}
		return s, nil
	case DiagramTypeBar:
		var s BarStyle
		if err := decodeStrict(raw, &s); err != nil {
			return nil, err
		}
		return s, nil
	case DiagramTypeLine:
		var s LineStyle
		if err := decodeStrict(raw, &s); err != nil {
			return nil, err
		}
		return s, nil

	default:
		return nil, fmt.Errorf("unsupported diagram type %q", string(t))
	}
}

// ============================================================
// Validation
// ============================================================

// Validate validates the view-only document.
func (vw ViewWindow) Validate() error {
	if len(vw.Diagrams) == 0 {
		return fmt.Errorf("diagrams must not be empty")
	}
	if err := validateWindowStyle("windowStyle", vw.WindowStyle); err != nil {
		return err
	}

	for i := range vw.Diagrams {
		if err := vw.Diagrams[i].Validate(i); err != nil {
			return fmt.Errorf("diagrams[%d]: %w", i, err)
		}
	}
	return nil
}

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

func (d Diagram) Validate(diagramIndex int) error {
	if d.Type == "" {
		return fmt.Errorf("type is required")
	}
	switch d.Type {
	case DiagramTypeBoolean, DiagramTypeBar, DiagramTypeLine:
		// ok
	default:
		return fmt.Errorf("type unsupported: %q", d.Type)
	}

	if len(d.Setup) == 0 {
		return fmt.Errorf("setup must not be empty")
	}

	for j, s := range d.Setup {
		prefix := fmt.Sprintf("setup[%d]", j)

		if strings.TrimSpace(s.Variable) == "" {
			return fmt.Errorf("%s.variable is required", prefix)
		}
		if strings.TrimSpace(s.Title) == "" {
			return fmt.Errorf("%s.title is required", prefix)
		}
		if s.VariableType == "" {
			return fmt.Errorf("%s.variableType is required", prefix)
		}

		if err := validateWindowStyle(prefix+".windowStyle", s.WindowStyle); err != nil {
			return err
		}

		if s.DiagramStyle != nil {
			switch d.Type {
			case DiagramTypeBoolean:
				style, ok := s.DiagramStyle.(BooleanStyle)
				if !ok {
					return fmt.Errorf("%s.diagramStyle must be boolean style", prefix)
				}
				if strings.TrimSpace(style.True) == "" || strings.TrimSpace(style.False) == "" {
					return fmt.Errorf("%s.diagramStyle requires non-empty keys \"true\" and \"false\"", prefix)
				}
				if s.VariableType != ValueTypeBoolean {
					return fmt.Errorf("%s.variableType must be %q for boolean diagram", prefix, ValueTypeBoolean)
				}

			case DiagramTypeBar:
				if _, ok := s.DiagramStyle.(BarStyle); !ok {
					return fmt.Errorf("%s.diagramStyle must be bar style", prefix)
				}
				if len(s.MultiVariableSetup) == 0 {
					if s.VariableType != ValueTypeInteger {
						return fmt.Errorf("%s.variableType must be %q for bar diagram", prefix, ValueTypeInteger)
					}
				}
			}
		}

		err := d.ValidateMultiVariable(prefix, &s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d Diagram) ValidateMultiVariable(prefix string, s *SetupItem) error {
	// Only validate multi-variable rules if the config actually uses it.
	if len(s.MultiVariableSetup) == 0 {
		return nil
	}

	// Must have at least 2 series.
	if len(s.MultiVariableSetup) < 2 {
		return fmt.Errorf("%s.multiVariableSetup must have at least 2 items", prefix)
	}

	// Only allow multi-series for bar/line (as you already decided).
	switch d.Type {
	case DiagramTypeBar, DiagramTypeLine:
		// ok
	default:
		return fmt.Errorf("%s.multiVariableSetup is only supported for bar/line diagrams", prefix)
	}

	// --- DiagramTypeBar: parent provides background; children provide bar color ---
	if d.Type == DiagramTypeBar {
		// Parent style must exist and be BarStyle so we can read background.
		if s.DiagramStyle == nil {
			return fmt.Errorf("%s.diagramStyle is required for multi-variable bar diagram", prefix)
		}

		parentStyle, ok := s.DiagramStyle.(BarStyle)
		if !ok {
			return fmt.Errorf("%s.diagramStyle must be bar style for multi-variable bar diagram", prefix)
		}

		if strings.TrimSpace(parentStyle.Background) == "" {
			return fmt.Errorf("%s.diagramStyle.background is required for multi-variable bar diagram", prefix)
		}

		// Optional but recommended: forbid setting parentStyle.Bar to avoid ambiguity.
		if strings.TrimSpace(parentStyle.Bar) != "" {
			return fmt.Errorf("%s.diagramStyle.bar must not be set on the parent in multi-variable bar diagram (set it per child)", prefix)
		}
	}

	// Validate children
	seen := make(map[string]struct{}, len(s.MultiVariableSetup))
	for k := range s.MultiVariableSetup {
		ms := &s.MultiVariableSetup[k]
		mp := fmt.Sprintf("%s.multiVariableSetup[%d]", prefix, k)

		if strings.TrimSpace(ms.Variable) == "" {
			return fmt.Errorf("%s.variable is required", mp)
		}
		if strings.TrimSpace(ms.Title) == "" {
			return fmt.Errorf("%s.title is required", mp)
		}
		if ms.VariableType == "" {
			return fmt.Errorf("%s.variableType is required", mp)
		}

		if _, dup := seen[ms.Variable]; dup {
			return fmt.Errorf("%s.variable %q is duplicated", mp, ms.Variable)
		}
		seen[ms.Variable] = struct{}{}

		switch d.Type {
		case DiagramTypeBar:
			// For multi-variable bar: each child MUST define bar color.
			if ms.VariableType != ValueTypeInteger {
				return fmt.Errorf("%s.variableType must be %q for bar diagram", mp, ValueTypeInteger)
			}

			if ms.DiagramStyle == nil {
				return fmt.Errorf("%s.diagramStyle is required for multi-variable bar diagram", mp)
			}

			childStyle, ok := ms.DiagramStyle.(BarStyle)
			if !ok {
				return fmt.Errorf("%s.diagramStyle must be bar style", mp)
			}

			if strings.TrimSpace(childStyle.Bar) == "" {
				return fmt.Errorf("%s.diagramStyle.bar is required for multi-variable bar diagram", mp)
			}

			// Optional: you can forbid child background/axis to keep responsibilities clean.
			// if strings.TrimSpace(childStyle.Background) != "" || strings.TrimSpace(childStyle.Axis) != "" {
			// 	return fmt.Errorf("%s.diagramStyle may only set \"bar\" for multi-variable bar diagram", mp)
			// }

		case DiagramTypeLine:
			// If you later want multi-line, define your required style rules here.
			// For now, you can keep it permissive or mirror your single-line rules.
		}
	}

	return nil
}
func validateWindowStyle(path string, ws *WindowStyle) error {
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
