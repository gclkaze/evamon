package port

type ChartRegistry struct {
	byID map[string]*DiagramUIRefs
}

func NewChartRegistry() *ChartRegistry {
	return &ChartRegistry{
		byID: make(map[string]*DiagramUIRefs),
	}
}

func (r ChartRegistry) Size() int {
	return len(r.byID)
}
func (r *ChartRegistry) Register(refs *DiagramUIRefs) {
	if refs == nil || refs.ID == "" {
		return
	}
	r.byID[refs.ID] = refs
}

func (r *ChartRegistry) Get(id string) (*DiagramUIRefs, bool) {
	refs, ok := r.byID[id]
	return refs, ok
}

func (r *ChartRegistry) Remove(id string) {
	delete(r.byID, id)
}
func (r *ChartRegistry) Exists(id string) bool {
	_, ok := r.byID[id]
	return ok
}
