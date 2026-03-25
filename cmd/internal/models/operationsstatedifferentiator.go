package models

import (
	"strings"
)

type ChangeType string

const (
	ChangeTypeRuleAdded         ChangeType = "rule_added"
	ChangeTypeRuleRemoved       ChangeType = "rule_removed"
	ChangeTypeLabelChanged      ChangeType = "label_changed"
	ChangeTypeExpressionChanged ChangeType = "expression_changed"
	ChangeTypeMaintainLink      ChangeType = "maintain_link"
	ChangeTypeActionsChanged    ChangeType = "actions_changed"
)

type snapshotRule struct {
	OriginalComponentID string
	Label               string
	Expression          string
	MaintainLink        bool
	ActionsKey          string
}

type RuleDiff struct {
	Rule                *TriggerRule
	LabelChanged        bool
	ExpressionChanged   bool
	MaintainLinkChanged bool
	ActionsChanged      bool
}

type DiffResult struct {
	AddedRules   []*TriggerRule
	RemovedRules []snapshotRule
	UpdatedRules []*RuleDiff
}

type OperationsStateDifferentiator struct {
	initialRules []snapshotRule
	initialMap   map[string]snapshotRule
	changes      map[ChangeType]bool
	hasChanged   bool
}

func NewOperationsStateDifferentiator(rules []*TriggerRule) *OperationsStateDifferentiator {
	snapshots := make([]snapshotRule, len(rules))
	initialMap := make(map[string]snapshotRule)
	for i, r := range rules {
		s := takeSnapshot(r)
		snapshots[i] = s
		initialMap[r.OriginalComponentID] = s
	}
	return &OperationsStateDifferentiator{
		initialRules: snapshots,
		initialMap:   initialMap,
		changes:      make(map[ChangeType]bool),
		hasChanged:   false,
	}
}

func takeSnapshot(r *TriggerRule) snapshotRule {
	paths := make([]string, len(r.Actions))
	for i, a := range r.Actions {
		paths[i] = a.Path
	}
	return snapshotRule{
		OriginalComponentID: r.OriginalComponentID,
		Label:               r.Label,
		Expression:          r.Expression,
		MaintainLink:        r.MaintainLink,
		ActionsKey:          strings.Join(paths, ","),
	}
}

func actionsKey(rule *TriggerRule) string {
	paths := make([]string, len(rule.Actions))
	for i, a := range rule.Actions {
		paths[i] = a.Path
	}
	return strings.Join(paths, ",")
}

func (d *OperationsStateDifferentiator) recompute() {
	for _, v := range d.changes {
		if v {
			d.hasChanged = true
			return
		}
	}
	d.hasChanged = false
}

func (d *OperationsStateDifferentiator) HasChanges() bool {
	return d.hasChanged
}

func (d *OperationsStateDifferentiator) GetDiff(current []*TriggerRule) DiffResult {
	result := DiffResult{
		AddedRules:   make([]*TriggerRule, 0),
		RemovedRules: make([]snapshotRule, 0),
		UpdatedRules: make([]*RuleDiff, 0),
	}

	currentMap := make(map[string]*TriggerRule)
	for _, r := range current {
		currentMap[r.OriginalComponentID] = r
	}

	for _, r := range current {
		if _, existed := d.initialMap[r.OriginalComponentID]; !existed {
			result.AddedRules = append(result.AddedRules, r)
		}
	}

	for _, s := range d.initialRules {
		if _, exists := currentMap[s.OriginalComponentID]; !exists {
			result.RemovedRules = append(result.RemovedRules, s)
		}
	}

	for _, r := range current {
		s, existed := d.initialMap[r.OriginalComponentID]
		if !existed {
			continue
		}
		diff := &RuleDiff{Rule: r}
		if r.Label != s.Label {
			diff.LabelChanged = true
		}
		if r.Expression != s.Expression {
			diff.ExpressionChanged = true
		}
		if r.MaintainLink != s.MaintainLink {
			diff.MaintainLinkChanged = true
		}
		if actionsKey(r) != s.ActionsKey {
			diff.ActionsChanged = true
		}
		if diff.LabelChanged || diff.ExpressionChanged || diff.MaintainLinkChanged || diff.ActionsChanged {
			result.UpdatedRules = append(result.UpdatedRules, diff)
		}
	}

	return result
}

func (d *OperationsStateDifferentiator) OnRuleAdded(current []*TriggerRule) {
	changed := false
	for _, r := range current {
		// empty OriginalComponentID = brand new rule, always counts as added
		if r.OriginalComponentID == "" {
			changed = true
			break
		}
		if _, existed := d.initialMap[r.OriginalComponentID]; !existed {
			changed = true
			break
		}
	}
	d.changes[ChangeTypeRuleAdded] = changed
	d.recompute()
}
func (d *OperationsStateDifferentiator) OnRuleRemoved(current []*TriggerRule) {
	currentMap := make(map[string]bool)
	for _, r := range current {
		if r.OriginalComponentID != "" {
			currentMap[r.OriginalComponentID] = true
		}
	}
	changed := false
	for _, s := range d.initialRules {
		if !currentMap[s.OriginalComponentID] {
			changed = true
			break
		}
	}
	d.changes[ChangeTypeRuleRemoved] = changed
	d.recompute()
}

func (d *OperationsStateDifferentiator) OnLabelChanged(current []*TriggerRule) {
	changed := false
	for _, r := range current {
		s, exists := d.initialMap[r.OriginalComponentID]
		if !exists {
			continue
		}
		if r.Label != s.Label {
			changed = true
			break
		}
	}
	d.changes[ChangeTypeLabelChanged] = changed
	d.recompute()
}

func (d *OperationsStateDifferentiator) OnExpressionChanged(current []*TriggerRule) {
	changed := false
	for _, r := range current {
		s, exists := d.initialMap[r.OriginalComponentID]
		if !exists {
			continue
		}
		if r.Expression != s.Expression {
			changed = true
			break
		}
	}
	d.changes[ChangeTypeExpressionChanged] = changed
	d.recompute()
}

func (d *OperationsStateDifferentiator) OnMaintainLinkChanged(current []*TriggerRule) {
	changed := false
	for _, r := range current {
		s, exists := d.initialMap[r.OriginalComponentID]
		if !exists {
			continue
		}
		if r.MaintainLink != s.MaintainLink {
			changed = true
			break
		}
	}
	d.changes[ChangeTypeMaintainLink] = changed
	d.recompute()
}

func (d *OperationsStateDifferentiator) OnActionsChanged(current []*TriggerRule) {
	changed := false
	for _, r := range current {
		s, exists := d.initialMap[r.OriginalComponentID]
		if !exists {
			continue
		}
		if actionsKey(r) != s.ActionsKey {
			changed = true
			break
		}
	}
	d.changes[ChangeTypeActionsChanged] = changed
	d.recompute()
}
