package viewproject

import "github.com/gclkaze/evamon/cmd/internal/ui/port"

type DiagramUIRefs struct {
	Panel   port.UIObject
	Toolbar port.UIObject
	Legend  port.UIObject
	Chart   port.UIObject
	Generic []port.UIObject
}

func NewDiagramUIRefs() *DiagramUIRefs {
	return &DiagramUIRefs{
		Generic: make([]port.UIObject, 0),
	}
}

func (r *DiagramUIRefs) AddGeneric(obj port.UIObject) {
	if r == nil || obj == nil {
		return
	}
	r.Generic = append(r.Generic, obj)
}
