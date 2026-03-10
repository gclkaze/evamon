package models

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
