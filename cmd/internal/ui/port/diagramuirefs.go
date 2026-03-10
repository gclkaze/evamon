package port

type DiagramUIRefs struct {
	ID      string
	Panel   UIObject
	Toolbar UIObject
	Legend  UIObject
	Chart   UIObject
	Generic []UIObject
}

func NewDiagramUIRefs() *DiagramUIRefs {
	return &DiagramUIRefs{
		Generic: make([]UIObject, 0),
	}
}

func (r *DiagramUIRefs) AddGeneric(obj UIObject) {
	if r == nil || obj == nil {
		return
	}
	r.Generic = append(r.Generic, obj)
}
