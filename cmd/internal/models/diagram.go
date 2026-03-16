package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gclkaze/evamon/pkg/utils"
)

// Diagram contains a typed union for setup items based on Diagram.Type.
type Diagram struct {
	ID    string       `json:"id"`
	Owner DiagramOwner `json:"-"`
	Type  DiagramType  `json:"type"`
	Setup []SetupItem  `json:"setup"`
}

func (d Diagram) GetName() string {
	if d.Setup == nil {
		return ""
	}

	if len(d.Setup) == 0 {
		return ""
	}

	return d.Setup[0].Title
}

func (d Diagram) GetFilter() *Filter {
	setups := d.GetSetup()
	if len(setups) == 0 {
		return nil
	}

	return setups[0].GetFilter()
}

func (d Diagram) GetID() string {
	return d.ID
}

func (d *Diagram) SetID(ID string) {
	d.ID = ID
}

type DiagramType string

const (
	DiagramTypeBoolean DiagramType = "boolean"
	DiagramTypeBar     DiagramType = "bar"
	DiagramTypeLine    DiagramType = "line"
)

type ValueType string

const (
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeInteger ValueType = "integer"
	ValueTypeFloat   ValueType = "float"
)

type setupItemRaw struct {
	Variable    string `json:"variable"`
	Title       string `json:"title"`
	Description string `json:"description"`

	VariableType ValueType       `json:"variableType"`
	DiagramStyle json.RawMessage `json:"diagramStyle,omitempty"`
	WindowStyle  *WindowStyle    `json:"windowStyle,omitempty"`

	Filter *Filter `json:"filter,omitempty"`

	MultiVariableSetup []json.RawMessage `json:"multiVariableSetup,omitempty"`
}

type multiSetupItemRaw struct {
	Variable    string `json:"variable"`
	Title       string `json:"title"`
	Description string `json:"description"`

	VariableType ValueType       `json:"variableType"`
	DiagramStyle json.RawMessage `json:"diagramStyle,omitempty"`
}

func (d Diagram) CollectVariables() []string {
	set := make(map[string]struct{})

	for _, s := range d.Setup {

		if v := strings.TrimSpace(s.Variable); v != "" {
			set[v] = struct{}{}
		}

		for _, ms := range s.MultiVariableSetup {
			if v := strings.TrimSpace(ms.Variable); v != "" {
				set[v] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(set))
	for v := range set {
		if !strings.HasPrefix(v, "$") {
			v = "$" + v
		}
		out = append(out, v)
	}

	return out
}
func (d *Diagram) GetDiagramOwner() DiagramOwner {
	if d == nil {
		return nil
	}
	return d.Owner
}

func (d *Diagram) GetProjectPath() string {
	if d == nil || d.Owner == nil {
		return ""
	}
	base := d.Owner.GetProjectBase()
	if base == nil {
		return ""
	}
	return base.ProjectPath
}

func (d *Diagram) GetSetup() []SetupItem {
	if d == nil {
		return nil
	}
	return d.Setup
}

func (d Diagram) GetType() DiagramType {
	return d.Type
}
func (d *Diagram) UnmarshalJSON(b []byte) error {
	// Step 1: decode diagram envelope (type + setup raw)
	var raw struct {
		ID    string            `json:"id"`
		Type  DiagramType       `json:"type"`
		Setup []json.RawMessage `json:"setup"`
	}
	if err := utils.DecodeStrict(b, &raw); err != nil {
		return err
	}

	if raw.Type == "" {
		return fmt.Errorf("diagram.type is required")
	}

	d.ID = raw.ID
	d.Type = raw.Type
	d.Setup = make([]SetupItem, 0, len(raw.Setup))

	for i, itemRaw := range raw.Setup {
		var item setupItemRaw
		if err := utils.DecodeStrict(itemRaw, &item); err != nil {
			return fmt.Errorf("diagram.setup[%d]: %w", i, err)
		}

		out := SetupItem{
			Variable:     item.Variable,
			Title:        item.Title,
			Description:  item.Description,
			VariableType: item.VariableType,
			WindowStyle:  item.WindowStyle,
			Filter:       item.Filter,
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
				if err := utils.DecodeStrict(subRaw, &sub); err != nil {
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

func (d Diagram) ValidateFilter(path string, s *SetupItem) error {
	if s == nil || s.Filter == nil {
		return nil
	}

	if s.Filter.Setup == nil {
		return fmt.Errorf("%s.setup is required when filter is present", path)
	}

	fs := s.Filter.Setup

	switch fs.Mode {
	case FilterModeAND, FilterModeOR:
		// ok
	default:
		return fmt.Errorf("%s.mode must be %q or %q", path+".setup", FilterModeAND, FilterModeOR)
	}

	if len(fs.Components) == 0 {
		return fmt.Errorf("%s.components must not be empty when filter is present", path+".setup")
	}

	seenIDs := make(map[string]struct{}, len(fs.Components))
	for i := range fs.Components {
		if strings.TrimSpace(fs.Components[i].ID) == "" {
			fs.Components[i].ID = utils.GetRandomString()
		}
	}
	for i, c := range fs.Components {
		cp := fmt.Sprintf("%s.components[%d]", path+".setup", i)

		if strings.TrimSpace(c.Expression) == "" {
			return fmt.Errorf("%s.expression is required", cp)
		}

		id := strings.TrimSpace(c.ID)
		if id == "" {
			continue
		}

		if _, exists := seenIDs[id]; exists {
			return fmt.Errorf("%s.id %q is duplicated", cp, id)
		}
		seenIDs[id] = struct{}{}
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

		if err := ValidateWindowStyle(prefix+".windowStyle", s.WindowStyle); err != nil {
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
		if err := d.ValidateFilter(prefix+".filter", &s); err != nil {
			return err
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
func parseStyleForDiagramType(t DiagramType, raw json.RawMessage) (Style, error) {
	switch t {
	case DiagramTypeBoolean:
		var s BooleanStyle
		if err := utils.DecodeStrict(raw, &s); err != nil {
			return nil, err
		}
		return s, nil
	case DiagramTypeBar:
		var s BarStyle
		if err := utils.DecodeStrict(raw, &s); err != nil {
			return nil, err
		}
		return s, nil
	case DiagramTypeLine:
		var s LineStyle
		if err := utils.DecodeStrict(raw, &s); err != nil {
			return nil, err
		}
		return s, nil

	default:
		return nil, fmt.Errorf("unsupported diagram type %q", string(t))
	}
}

func EnsureDiagramIDsInViewWindow(vw *ViewWindow, seen map[string]struct{}) bool {
	if vw == nil {
		return false
	}

	modified := false

	for i := range vw.Diagrams {
		id := vw.Diagrams[i].GetID()

		if id == "" {
			id = newUniqueDiagramID(seen)
			vw.Diagrams[i].SetID(id)
			modified = true
		} else if _, exists := seen[id]; exists {
			id = newUniqueDiagramID(seen)
			vw.Diagrams[i].SetID(id)
			modified = true
		}

		seen[id] = struct{}{}
	}

	return modified
}
