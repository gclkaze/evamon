package utils

import (
	"fmt"
	"slices"
	"strings"

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
	res := conditionParser.Analyze(expr)
	return res
}

func AnalyzeExpressionWithRespectToTheDeclaredVariables(variables []string, expr string) error {
	conditionParser := &tafexpr.TAFArgumentParser{}
	conditionParser.VariableContext = SetupTruthyVariableContext()
	res := conditionParser.Analyze(expr)
	if !res {
		return parserErrorsToError(conditionParser)
	}

	foundExprs := conditionParser.VariableExpressions
	//lets see the relevance with our variables from the chart
	for i := range foundExprs {
		expr := foundExprs[i]
		name := expr.VarName
		found := slices.Contains(variables, name)

		if !found {
			return fmt.Errorf("undeclared chart variable '%s'", name)
		}
	}
	return nil
}

func parserErrorsToError(parser *tafexpr.TAFArgumentParser) error {
	errors := parser.GetErrors()
	if len(errors) != 0 {
		return fmt.Errorf("%s", strings.Join(errors, ". "))
	}
	return nil
}
