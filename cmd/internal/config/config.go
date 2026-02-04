package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// --------------------
// Public model
// --------------------

type Config struct {
	// Absolute path of the config file (filled by loader)
	ConfigPath string `json:"-"`

	// Script path as given in JSON (raw), and the resolved absolute path.
	Script    string `json:"script"`
	ScriptAbs string `json:"-"`

	View View `json:"view"`
}

type View struct {
	MultiTab    bool         `json:"multiTab"`
	WindowStyle *WindowStyle `json:"windowStyle,omitempty"`
	Diagrams    []Diagram    `json:"diagrams"`
}

type WindowStyle struct {
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
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
)

// SetupItem is one variable/series to visualize.
type SetupItem struct {
	Variable     string       `json:"variable"`
	Title        string       `json:"title"`
	VariableType ValueType    `json:"variableType"`
	WindowStyle  *WindowStyle `json:"windowStyle,omitempty"`

	// Typed union:
	// - for boolean diagrams: *BooleanStyle
	// - for bar diagrams:     *BarStyle
	DiagramStyle Style `json:"diagramStyle,omitempty"`
}

type ValueType string

const (
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeInteger ValueType = "integer"
)

// --------------------
// Typed union for diagramStyle
// --------------------

type Style interface {
	isStyle()
}

// BooleanStyle matches your example keys "true"/"false".
type BooleanStyle struct {
	True  string `json:"true"`
	False string `json:"false"`
}

func (BooleanStyle) isStyle() {}

// BarStyle matches your example keys "axis"/"background".
type BarStyle struct {
	Axis       string `json:"axis,omitempty"`
	Background string `json:"background,omitempty"`
}

func (BarStyle) isStyle() {}

// --------------------
// Loader
// --------------------

// Load reads a JSON config file from disk, parses it, resolves ScriptAbs,
// and validates the result.
func Load(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	expanded, err := expandUser(path)
	if err != nil {
		return nil, err
	}

	absCfgPath, err := filepath.Abs(expanded)
	if err != nil {
		return nil, fmt.Errorf("abs(%q): %w", expanded, err)
	}

	data, err := os.ReadFile(absCfgPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", absCfgPath, err)
	}

	var cfg Config
	if err := decodeStrict(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %q: %w", absCfgPath, err)
	}

	cfg.ConfigPath = absCfgPath

	// Resolve script relative to config location (if not absolute).
	if strings.TrimSpace(cfg.Script) == "" {
		return nil, fmt.Errorf("invalid config %q: script is required", absCfgPath)
	}

	scriptExpanded, err := expandUser(cfg.Script)
	if err != nil {
		return nil, fmt.Errorf("invalid config %q: %w", absCfgPath, err)
	}

	cfgDir := filepath.Dir(absCfgPath)
	if filepath.IsAbs(scriptExpanded) {
		cfg.ScriptAbs = scriptExpanded
	} else {
		cfg.ScriptAbs = filepath.Join(cfgDir, scriptExpanded)
	}

	// Normalize script abs path (clean + abs)
	cfg.ScriptAbs, err = filepath.Abs(cfg.ScriptAbs)
	if err != nil {
		return nil, fmt.Errorf("invalid config %q: abs(script): %w", absCfgPath, err)
	}

	// Validate everything
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %q: %w", absCfgPath, err)
	}

	return &cfg, nil
}

// decodeStrict decodes JSON and fails on unknown fields.
func decodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	// Optional: prevent numbers silently turning into float64 in interface{} cases.
	// (We don't rely on interface{} here, but it doesn't hurt.)
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

// --------------------
// Custom unmarshalling for typed diagramStyle union
// --------------------

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

	// Step 2: decode each setup item, interpreting diagramStyle based on d.Type
	for i, itemRaw := range raw.Setup {
		var item setupItemRaw
		if err := decodeStrict(itemRaw, &item); err != nil {
			return fmt.Errorf("diagram.setup[%d]: %w", i, err)
		}

		out := SetupItem{
			Variable:     item.Variable,
			Title:        item.Title,
			VariableType: item.VariableType,
			WindowStyle:  item.WindowStyle,
		}

		if len(item.DiagramStyle) > 0 && string(item.DiagramStyle) != "null" {
			style, err := parseStyleForDiagramType(raw.Type, item.DiagramStyle)
			if err != nil {
				return fmt.Errorf("diagram.setup[%d].diagramStyle: %w", i, err)
			}
			out.DiagramStyle = style
		}

		d.Setup = append(d.Setup, out)
	}

	return nil
}

// Internal helper type for decoding setup items with raw diagramStyle.
type setupItemRaw struct {
	Variable     string          `json:"variable"`
	Title        string          `json:"title"`
	VariableType ValueType       `json:"variableType"`
	DiagramStyle json.RawMessage `json:"diagramStyle,omitempty"`
	WindowStyle  *WindowStyle    `json:"windowStyle,omitempty"`
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
	default:
		return nil, fmt.Errorf("unsupported diagram type %q", string(t))
	}
}

// --------------------
// Validation
// --------------------

func (c Config) Validate() error {
	if strings.TrimSpace(c.Script) == "" {
		return fmt.Errorf("script is required")
	}
	if strings.TrimSpace(c.ScriptAbs) == "" {
		return fmt.Errorf("scriptAbs is empty (loader bug)")
	}

	// Validate script exists
	if st, err := os.Stat(c.ScriptAbs); err != nil {
		return fmt.Errorf("script not found: %q (%v)", c.ScriptAbs, err)
	} else if st.IsDir() {
		return fmt.Errorf("script path is a directory: %q", c.ScriptAbs)
	}

	// Validate view
	if len(c.View.Diagrams) == 0 {
		return fmt.Errorf("view.diagrams must not be empty")
	}
	if err := validateWindowStyle("view.windowStyle", c.View.WindowStyle); err != nil {
		return err
	}

	for i := range c.View.Diagrams {
		if err := c.View.Diagrams[i].Validate(i); err != nil {
			return err
		}
	}
	return nil
}

func (d Diagram) Validate(diagramIndex int) error {
	if d.Type == "" {
		return fmt.Errorf("view.diagrams[%d].type is required", diagramIndex)
	}
	switch d.Type {
	case DiagramTypeBoolean, DiagramTypeBar:
		// ok
	default:
		return fmt.Errorf("view.diagrams[%d].type unsupported: %q", diagramIndex, d.Type)
	}

	if len(d.Setup) == 0 {
		return fmt.Errorf("view.diagrams[%d].setup must not be empty", diagramIndex)
	}

	for j, s := range d.Setup {
		prefix := fmt.Sprintf("view.diagrams[%d].setup[%d]", diagramIndex, j)

		if strings.TrimSpace(s.Variable) == "" {
			return fmt.Errorf("%s.variable is required", prefix)
		}
		if strings.TrimSpace(s.Title) == "" {
			return fmt.Errorf("%s.title is required", prefix)
		}
		if s.VariableType == "" {
			return fmt.Errorf("%s.variableType is required", prefix)
		}

		// windowStyle ints must be positive if set
		if err := validateWindowStyle(prefix+".windowStyle", s.WindowStyle); err != nil {
			return err
		}

		// Typed style validation by diagram type
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
				// Optional: enforce variableType is boolean
				if s.VariableType != ValueTypeBoolean {
					return fmt.Errorf("%s.variableType must be %q for boolean diagram", prefix, ValueTypeBoolean)
				}

			case DiagramTypeBar:
				_, ok := s.DiagramStyle.(BarStyle)
				if !ok {
					return fmt.Errorf("%s.diagramStyle must be bar style", prefix)
				}
				// Optional: enforce integer for bars
				if s.VariableType != ValueTypeInteger {
					return fmt.Errorf("%s.variableType must be %q for bar diagram", prefix, ValueTypeInteger)
				}
			}
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
