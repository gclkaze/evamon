package utils

import (
	tafexpr "github.com/gclkaze/tafexpr"
	vc "github.com/gclkaze/tafexpr/variablecontext"
)

func SetupTruthyVariableContext() vc.IVariableContext {
	v := &vc.TruthyVariablecontext{}
	v.Init(true)

	var context vc.IVariableContext = v
	return context
}

func AnalyzeExpression(expr string) bool {
	conditionParser := &tafexpr.TAFArgumentParser{}
	conditionParser.VariableContext = SetupTruthyVariableContext()

	return conditionParser.Analyze(expr)
}
