package models

import (
	"testing"
)

func rule(id, label, expr string, maintainLink bool, actions ...string) *TriggerRule {
	acts := make([]*ActionFile, len(actions))
	for i, p := range actions {
		acts[i] = &ActionFile{Path: p}
	}
	return &TriggerRule{
		OriginalComponentID: id,
		Label:               label,
		Expression:          expr,
		MaintainLink:        maintainLink,
		Actions:             acts,
	}
}

func TestNewDifferentiator_NoChanges(t *testing.T) {
	rules := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(rules)

	if d.HasChanges() {
		t.Fatal("expected no changes on fresh differentiator")
	}
}

func TestOnRuleAdded_DetectsNewRule(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", true), rule("b", "B", "y>0", false)}
	d.OnRuleAdded(current)

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after adding a rule")
	}
}

func TestOnRuleAdded_NoNewRule(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	d.OnRuleAdded(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when no rule was added")
	}
}

func TestOnRuleRemoved_DetectsRemovedRule(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true), rule("b", "B", "y>0", false)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", true)}
	d.OnRuleRemoved(current)

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after removing a rule")
	}
}

func TestOnRuleRemoved_NothingRemoved(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	d.OnRuleRemoved(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when nothing was removed")
	}
}

func TestOnLabelChanged_DetectsChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A-modified", "x>0", true)}
	d.OnLabelChanged(current)

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after label change")
	}
}

func TestOnLabelChanged_NoChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	d.OnLabelChanged(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when label unchanged")
	}
}

func TestOnLabelChanged_SkipsUnknownRules(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	// "b" is not in initial map, so label change on "b" should not count
	current := []*TriggerRule{rule("b", "B-changed", "y>0", true)}
	d.OnLabelChanged(current)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false for unknown rule IDs")
	}
}

func TestOnExpressionChanged_DetectsChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>1", true)}
	d.OnExpressionChanged(current)

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after expression change")
	}
}

func TestOnExpressionChanged_NoChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	d.OnExpressionChanged(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when expression unchanged")
	}
}

func TestOnMaintainLinkChanged_DetectsChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", false)}
	d.OnMaintainLinkChanged(current)

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after MaintainLink change")
	}
}

func TestOnMaintainLinkChanged_NoChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	d.OnMaintainLinkChanged(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when MaintainLink unchanged")
	}
}

func TestOnActionsChanged_DetectsChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true, "/script1.sh")}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", true, "/script1.sh", "/script2.sh")}
	d.OnActionsChanged(current)

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after actions change")
	}
}

func TestOnActionsChanged_NoChange(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true, "/script1.sh")}
	d := NewOperationsStateDifferentiator(initial)

	d.OnActionsChanged(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when actions unchanged")
	}
}

func TestOnActionsChanged_EmptyToEmpty(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	d.OnActionsChanged(initial)

	if d.HasChanges() {
		t.Fatal("expected HasChanges=false when both have no actions")
	}
}

func TestHasChanges_RevertedChangeBecomesClean(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	modified := []*TriggerRule{rule("a", "A-modified", "x>0", true)}
	d.OnLabelChanged(modified)
	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true after label change")
	}

	// Revert label back to original
	d.OnLabelChanged(initial)
	if d.HasChanges() {
		t.Fatal("expected HasChanges=false after reverting label")
	}
}

func TestHasChanges_MultipleChangeTypes(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	// Change label
	d.OnLabelChanged([]*TriggerRule{rule("a", "A-modified", "x>0", true)})
	// Also add a rule
	d.OnRuleAdded([]*TriggerRule{rule("a", "A-modified", "x>0", true), rule("b", "B", "y>0", false)})

	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true with multiple change types active")
	}

	// Revert label but keep the added rule
	d.OnLabelChanged([]*TriggerRule{rule("a", "A", "x>0", true), rule("b", "B", "y>0", false)})
	if !d.HasChanges() {
		t.Fatal("expected HasChanges=true because added rule is still present")
	}
}

// --- GetDiff ---

func TestGetDiff_AddedRule(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", true), rule("b", "B", "y>0", false)}
	diff := d.GetDiff(current)

	if len(diff.AddedRules) != 1 || diff.AddedRules[0].OriginalComponentID != "b" {
		t.Fatalf("expected 1 added rule 'b', got %+v", diff.AddedRules)
	}
	if len(diff.RemovedRules) != 0 {
		t.Fatalf("expected 0 removed rules, got %d", len(diff.RemovedRules))
	}
	if len(diff.UpdatedRules) != 0 {
		t.Fatalf("expected 0 updated rules, got %d", len(diff.UpdatedRules))
	}
}

func TestGetDiff_RemovedRule(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true), rule("b", "B", "y>0", false)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", true)}
	diff := d.GetDiff(current)

	if len(diff.RemovedRules) != 1 || diff.RemovedRules[0].OriginalComponentID != "b" {
		t.Fatalf("expected 1 removed rule 'b', got %+v", diff.RemovedRules)
	}
	if len(diff.AddedRules) != 0 {
		t.Fatalf("expected 0 added rules, got %d", len(diff.AddedRules))
	}
}

func TestGetDiff_UpdatedLabel(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A-changed", "x>0", true)}
	diff := d.GetDiff(current)

	if len(diff.UpdatedRules) != 1 {
		t.Fatalf("expected 1 updated rule, got %d", len(diff.UpdatedRules))
	}
	rd := diff.UpdatedRules[0]
	if !rd.LabelChanged {
		t.Error("expected LabelChanged=true")
	}
	if rd.ExpressionChanged || rd.MaintainLinkChanged || rd.ActionsChanged {
		t.Error("expected only LabelChanged to be true")
	}
}

func TestGetDiff_UpdatedExpression(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>99", true)}
	diff := d.GetDiff(current)

	if len(diff.UpdatedRules) != 1 || !diff.UpdatedRules[0].ExpressionChanged {
		t.Fatal("expected ExpressionChanged=true")
	}
}

func TestGetDiff_UpdatedMaintainLink(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", false)}
	diff := d.GetDiff(current)

	if len(diff.UpdatedRules) != 1 || !diff.UpdatedRules[0].MaintainLinkChanged {
		t.Fatal("expected MaintainLinkChanged=true")
	}
}

func TestGetDiff_UpdatedActions(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true, "/s1.sh")}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A", "x>0", true, "/s1.sh", "/s2.sh")}
	diff := d.GetDiff(current)

	if len(diff.UpdatedRules) != 1 || !diff.UpdatedRules[0].ActionsChanged {
		t.Fatal("expected ActionsChanged=true")
	}
}

func TestGetDiff_NoChanges(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true, "/s1.sh")}
	d := NewOperationsStateDifferentiator(initial)

	diff := d.GetDiff(initial)

	if len(diff.AddedRules) != 0 || len(diff.RemovedRules) != 0 || len(diff.UpdatedRules) != 0 {
		t.Fatalf("expected empty diff, got added=%d removed=%d updated=%d",
			len(diff.AddedRules), len(diff.RemovedRules), len(diff.UpdatedRules))
	}
}

func TestGetDiff_MultipleFieldsChanged(t *testing.T) {
	initial := []*TriggerRule{rule("a", "A", "x>0", true)}
	d := NewOperationsStateDifferentiator(initial)

	current := []*TriggerRule{rule("a", "A-new", "x>1", false)}
	diff := d.GetDiff(current)

	if len(diff.UpdatedRules) != 1 {
		t.Fatalf("expected 1 updated rule, got %d", len(diff.UpdatedRules))
	}
	rd := diff.UpdatedRules[0]
	if !rd.LabelChanged || !rd.ExpressionChanged || !rd.MaintainLinkChanged {
		t.Errorf("expected Label, Expression, MaintainLink all changed: %+v", rd)
	}
}

func TestGetDiff_EmptyInitialAndCurrent(t *testing.T) {
	d := NewOperationsStateDifferentiator([]*TriggerRule{})

	diff := d.GetDiff([]*TriggerRule{})

	if len(diff.AddedRules) != 0 || len(diff.RemovedRules) != 0 || len(diff.UpdatedRules) != 0 {
		t.Fatal("expected all-empty diff for empty initial and current")
	}
}
