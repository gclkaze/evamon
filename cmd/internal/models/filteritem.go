package models

type FilterItem struct {
	Label      string
	Expression string

	SourceComponentID string
	MaintainLink      bool

	Files []*ActionFile
}

func NewFilterItem(source FilterComponent) *FilterItem {
	return &FilterItem{
		SourceComponentID: source.ID,
		MaintainLink:      true,
		Files:             make([]*ActionFile, 0),
	}
}

func (fi *FilterItem) ResolvedLabel(components []FilterComponent) string {
	if fi.MaintainLink && fi.SourceComponentID != "" {
		for _, c := range components {
			if c.ID == fi.SourceComponentID {
				return c.Label
			}
		}
	}
	return fi.Label
}

func (fi *FilterItem) ResolvedExpression(components []FilterComponent) string {
	if fi.MaintainLink && fi.SourceComponentID != "" {
		for _, c := range components {
			if c.ID == fi.SourceComponentID {
				return c.Expression
			}
		}
	}
	return fi.Expression
}

func (fi *FilterItem) BreakLink(components []FilterComponent) {
	if !fi.MaintainLink && fi.SourceComponentID != "" {
		for _, c := range components {
			if c.ID == fi.SourceComponentID {
				fi.Label = c.Label
				fi.Expression = c.Expression
				break
			}
		}
		fi.SourceComponentID = ""
	}
}

func (fi *FilterItem) HasFile(path string) bool {
	for _, f := range fi.Files {
		if f.Path == path {
			return true
		}
	}
	return false
}

func (fi *FilterItem) AddFile(file *ActionFile) {
	if !fi.HasFile(file.Path) {
		fi.Files = append(fi.Files, file)
	}
}

func (fi *FilterItem) RemoveFile(path string) {
	filtered := fi.Files[:0]
	for _, f := range fi.Files {
		if f.Path != path {
			filtered = append(filtered, f)
		}
	}
	fi.Files = filtered
}
