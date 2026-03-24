package operations

import "fyne.io/fyne/v2"

type threeColLayoutStruct struct {
	leftWidth  float32
	midWidth   float32
	rightWidth float32
}

func threeColLayout(left, mid, right float32) fyne.Layout {
	return &threeColLayoutStruct{
		leftWidth:  left,
		midWidth:   mid,
		rightWidth: right,
	}
}

func (t *threeColLayoutStruct) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	objects[0].Resize(fyne.NewSize(t.leftWidth, size.Height))
	objects[0].Move(fyne.NewPos(0, 0))

	objects[1].Resize(fyne.NewSize(t.midWidth, size.Height))
	objects[1].Move(fyne.NewPos(t.leftWidth, 0))

	objects[2].Resize(fyne.NewSize(t.rightWidth, size.Height))
	objects[2].Move(fyne.NewPos(t.leftWidth+t.midWidth, 0))
}

func (t *threeColLayoutStruct) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(t.leftWidth+t.midWidth+t.rightWidth, 0)
}
