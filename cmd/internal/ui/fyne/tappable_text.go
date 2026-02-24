package fynerenderer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type TapOverlay struct {
	widget.BaseWidget
	OnTapped func()
}

func NewTapOverlay(onTapped func()) *TapOverlay {
	o := &TapOverlay{OnTapped: onTapped}
	o.ExtendBaseWidget(o)
	return o
}

func (o *TapOverlay) CreateRenderer() fyne.WidgetRenderer {
	// Nothing visible; it just captures taps.
	r := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	return widget.NewSimpleRenderer(r)
}

func (o *TapOverlay) Tapped(*fyne.PointEvent) {
	if o.OnTapped != nil {
		o.OnTapped()
	}
}

func (o *TapOverlay) TappedSecondary(*fyne.PointEvent) {}
