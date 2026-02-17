package fynediagrams

import (
	"image/color"

	"github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
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

/*
	func (f *Factory) NewBoolFill(opts port.BoolFillOptions) port.Drawer {
		trueC := opts.TrueColor
		falseC := opts.FalseColor
		if trueC == nil {
			trueC = f.defaultTrue
		}
		if falseC == nil {
			falseC = f.defaultFalse
		}
		return NewBoolFillDrawer(trueC, falseC)
	}

	func (f *Factory) NewBarChart(opts port.BarChartOptions) port.Drawer {
		return NewBarChartDrawer(opts.MaxPoints, opts.Width, opts.Height, opts.Axis, opts.Background)
	}
*/
func (f *Factory) NewBoolFill(opts port.BoolFillOptions) port.DiagramWidget {
	trueC := opts.TrueColor
	falseC := opts.FalseColor
	if trueC == nil {
		trueC = f.defaultTrue
	}
	if falseC == nil {
		falseC = f.defaultFalse
	}

	drawer := NewBoolFillDrawer(trueC, falseC)
	return NewBoolFillWidget(drawer) // tiny wrapper to satisfy DiagramWidget
}

func (f *Factory) NewBarChart(opts port.BarChartOptions) port.DiagramWidget {
	drawer := NewBarChartDrawer(opts.MaxPoints, opts.Width, opts.Height, opts.Axis, opts.Background)
	return NewBarChartWidget(drawer) // the resize-aware wrapper
}
