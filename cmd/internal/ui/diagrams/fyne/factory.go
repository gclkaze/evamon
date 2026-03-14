package fynediagrams

import (
	"image/color"

	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	"github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
	dia "github.com/gclkaze/evamon/cmd/internal/ui/port"
)

type Factory struct {
	defaultTrue  color.Color
	defaultFalse color.Color
}

func NewFactory() *Factory {
	return &Factory{
		defaultTrue:  color.NRGBA{R: 0, G: 180, B: 0, A: 255},
		defaultFalse: color.NRGBA{R: 200, G: 0, B: 0, A: 255},
	}
}

func (f *Factory) NewBoolFill(opts port.BoolFillOptions, title string, description string, width, height float32, varname string /*, owner *models.SetupItem*/) port.DiagramWidget {
	trueC := opts.TrueColor
	falseC := opts.FalseColor
	if trueC == nil {
		trueC = f.defaultTrue
	}
	if falseC == nil {
		falseC = f.defaultFalse
	}

	drawer := NewBoolFillDrawer(trueC, falseC, varname)
	return NewBoolFillWidget(drawer, title, description, width, height) // tiny wrapper to satisfy DiagramWidget
}

func (f *Factory) NewBarChart(opts port.BarChartOptions, title string, description string, width, height float32, owner dia.IDiagram) port.DiagramWidget {
	seriesCount := len(opts.Variables)
	if seriesCount <= 0 {
		seriesCount = 1
	}

	src := data.NewMultiSeriesRing(
		opts.MaxPoints, // history size
		seriesCount,    // number of variables
		owner,
		collectVariableNames(opts.Variables),
	)
	drawer := NewBarChartDrawer(src, opts.Width, opts.Height, opts.Variables, opts.Background)
	return NewBarChartWidget(drawer, title, description, width, height) // the resize-aware wrapper
}

func (f *Factory) NewLineChart(opts port.LineChartOptions, title string, description string, initialWidth, initialHeight, width, height float32, owner dia.IDiagram) port.DiagramWidget {
	src := data.NewMultiSeriesRing(opts.MaxPoints, len(opts.Variables), owner, collectVariableNames(opts.Variables)) // example impl
	drawer := NewLineChartDrawer(src, opts, initialWidth, initialHeight)
	return NewLineChartWidget(drawer, title, description, width, height)
}

func collectVariableNames(vs []port.VariableStyle) []string {
	var vars []string
	for i := range vs {
		vars = append(vars, vs[i].VariableName)
	}

	return vars
}
