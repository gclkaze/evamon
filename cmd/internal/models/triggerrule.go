package models

type TriggerRule struct {
	Label      string `json:"label,omitempty"`
	Expression string `json:"expression,omitempty"`

	SourceComponentID   string `json:"sourceComponentID,omitempty"`
	OriginalComponentID string `json:"originalComponentID"`
	MaintainLink        bool   `json:"maintainLink"`
	Edited              bool   `json:"edited"`

	Actions []*ActionFile `json:"actions,omitempty"`
}

func NewTriggerRule(source FilterComponent) *TriggerRule {
	return &TriggerRule{
		SourceComponentID:   source.ID,
		OriginalComponentID: source.ID,
		MaintainLink:        true,
		Edited:              false,
		Actions:             make([]*ActionFile, 0),
	}
}

func (tr *TriggerRule) ResolvedLabel(components []FilterComponent) string {
	if tr.MaintainLink && tr.SourceComponentID != "" {
		for _, c := range components {
			if c.ID == tr.SourceComponentID {
				return c.Label
			}
		}
	}
	return tr.Label
}

func (tr *TriggerRule) ResolvedExpression(components []FilterComponent) string {
	if tr.MaintainLink && tr.SourceComponentID != "" {
		for _, c := range components {
			if c.ID == tr.SourceComponentID {
				return c.Expression
			}
		}
	}
	return tr.Expression
}

func (tr *TriggerRule) BreakLink(components []FilterComponent) {
	if !tr.MaintainLink && tr.SourceComponentID != "" {
		for _, c := range components {
			if c.ID == tr.SourceComponentID {
				tr.Label = c.Label
				tr.Expression = c.Expression
				break
			}
		}
		tr.SourceComponentID = ""
	}
}

func (tr *TriggerRule) RestoreLink() {
	tr.MaintainLink = true
	tr.SourceComponentID = tr.OriginalComponentID
	tr.Label = ""
	tr.Expression = ""
}

func (tr *TriggerRule) AddAction(file *ActionFile) {
	tr.Actions = append(tr.Actions, file)
}

func (tr *TriggerRule) RemoveActionAt(index int) {
	if index < 0 || index >= len(tr.Actions) {
		return
	}
	tr.Actions = append(tr.Actions[:index], tr.Actions[index+1:]...)
}

func (tr *TriggerRule) MoveActionUp(index int) {
	if index <= 0 || index >= len(tr.Actions) {
		return
	}
	tr.Actions[index-1], tr.Actions[index] = tr.Actions[index], tr.Actions[index-1]
}

func (tr *TriggerRule) MoveActionDown(index int) {
	if index < 0 || index >= len(tr.Actions)-1 {
		return
	}
	tr.Actions[index], tr.Actions[index+1] = tr.Actions[index+1], tr.Actions[index]
}
