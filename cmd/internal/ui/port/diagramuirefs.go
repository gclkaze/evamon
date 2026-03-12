package port

type DiagramUIRefs struct {
	ID      string
	Panel   UIObject
	Toolbar UIObject
	Legend  UIObject
	Chart   UIObject
	Generic []UIObject
	parts   map[string]UIObject

	rebuildToolbar func() []UIObject

	renderer Renderer
}

const (
	partMainChart             = "mainchart"
	partToolbar               = "toolbar"
	partLegend                = "legend"
	partMaximizeButton        = "maximize_button"
	partDownloadButton        = "download_button"
	partFilterEnabledCheckbox = "filter_enabled_checkbox"
	partFilterButton          = "filter_button"
)

func NewDiagramUIRefs(id string, renderer Renderer) *DiagramUIRefs {
	return &DiagramUIRefs{
		ID:       id,
		Generic:  make([]UIObject, 0),
		parts:    make(map[string]UIObject),
		renderer: renderer,
	}
}

func (r *DiagramUIRefs) AddGeneric(obj UIObject) {
	if r == nil || obj == nil {
		return
	}
	r.Generic = append(r.Generic, obj)
}

// internal helpers
func (r *DiagramUIRefs) set(suffix string, obj UIObject) {
	key := r.ID + "_" + suffix
	r.parts[key] = obj
}

func (r *DiagramUIRefs) get(suffix string) (UIObject, bool) {
	key := r.ID + "_" + suffix
	obj, ok := r.parts[key]
	return obj, ok
}

func (r *DiagramUIRefs) SetRebuildToolbar(fn func() []UIObject) {
	r.rebuildToolbar = fn
}

func (r *DiagramUIRefs) RebuildToolbar() {
	if r.rebuildToolbar == nil {
		return
	}
	toolbar, ok := r.GetToolbar()
	if !ok {
		return
	}
	newChildren := r.rebuildToolbar()
	r.renderer.Layout().ReplaceHBoxContent(toolbar, newChildren...)
}

// MainChart
func (r *DiagramUIRefs) RegisterMainChart(obj UIObject) { r.set(partMainChart, obj) }
func (r *DiagramUIRefs) GetMainChart() (UIObject, bool) { return r.get(partMainChart) }
func (r *DiagramUIRefs) ReplaceMainChart(obj UIObject)  { r.set(partMainChart, obj) }

// Toolbar
func (r *DiagramUIRefs) RegisterToolbar(obj UIObject) { r.set(partToolbar, obj) }
func (r *DiagramUIRefs) GetToolbar() (UIObject, bool) { return r.get(partToolbar) }
func (r *DiagramUIRefs) ReplaceToolbar(obj UIObject)  { r.set(partToolbar, obj) }

// Legend
func (r *DiagramUIRefs) RegisterLegend(obj UIObject) { r.set(partLegend, obj) }
func (r *DiagramUIRefs) GetLegend() (UIObject, bool) { return r.get(partLegend) }
func (r *DiagramUIRefs) ReplaceLegend(obj UIObject)  { r.set(partLegend, obj) }

// MaximizeButton
func (r *DiagramUIRefs) RegisterMaximizeButton(obj UIObject) { r.set(partMaximizeButton, obj) }
func (r *DiagramUIRefs) GetMaximizeButton() (UIObject, bool) { return r.get(partMaximizeButton) }
func (r *DiagramUIRefs) ReplaceMaximizeButton(obj UIObject)  { r.set(partMaximizeButton, obj) }

// DownloadButton
func (r *DiagramUIRefs) RegisterDownloadButton(obj UIObject) { r.set(partDownloadButton, obj) }
func (r *DiagramUIRefs) GetDownloadButton() (UIObject, bool) { return r.get(partDownloadButton) }
func (r *DiagramUIRefs) ReplaceDownloadButton(obj UIObject)  { r.set(partDownloadButton, obj) }

// FilterEnabledCheckbox
func (r *DiagramUIRefs) RegisterFilterEnabledCheckbox(obj UIObject) {
	r.set(partFilterEnabledCheckbox, obj)
}
func (r *DiagramUIRefs) GetFilterEnabledCheckbox() (UIObject, bool) {
	return r.get(partFilterEnabledCheckbox)
}
func (r *DiagramUIRefs) ReplaceFilterEnabledCheckbox(obj UIObject) {
	r.set(partFilterEnabledCheckbox, obj)
}

// Filter Button
func (r *DiagramUIRefs) RegisterFilterButton(obj UIObject) {
	r.set(partFilterButton, obj)
}
func (r *DiagramUIRefs) GetFilterButton() (UIObject, bool) { return r.get(partFilterButton) }
func (r *DiagramUIRefs) ReplaceFilterButton(obj UIObject)  { r.set(partFilterButton, obj) }
