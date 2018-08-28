package evaluator

import (
	"github.com/SealNTibbers/GotalkInterpreter/parser"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

//testing stuff
func NewTestEvaluator() *Evaluator {
	evaluator := new(Evaluator)
	evaluator.globalScope = new(treeNodes.Scope).Initialize()
	return evaluator
}

func TestEval(codeString string) treeNodes.SmalltalkObjectInterface {
	evaluator := NewTestEvaluator()
	programNode := parser.InitializeParserFor(codeString)
	return evaluator.EvaluateProgramNode(programNode)
}

func TestEvalWithScope(codeString string, scope *treeNodes.Scope) treeNodes.SmalltalkObjectInterface {
	evaluator := NewEvaluatorWithGlobalScope(scope)
	programNode := parser.InitializeParserFor(codeString)
	return evaluator.EvaluateProgramNode(programNode)
}

//real world API
func NewSmalltalkVM() *Evaluator {
	globalScope := new(treeNodes.Scope).Initialize()
	return NewEvaluatorWithGlobalScope(globalScope)
}

func NewEvaluatorWithGlobalScope(global *treeNodes.Scope) *Evaluator {
	evaluator := new(Evaluator)
	evaluator.programCache = make(map[string]treeNodes.ProgramNodeInterface)
	evaluator.globalScope = global
	return evaluator
}

type Evaluator struct {
	globalScope  *treeNodes.Scope
	programCache map[string]treeNodes.ProgramNodeInterface
}

func (e *Evaluator) SetGlobalScope(scope *treeNodes.Scope) *Evaluator {
	e.globalScope = scope
	return e
}

func (e *Evaluator) GetGlobalScope() *treeNodes.Scope {
	return e.globalScope
}

func (e *Evaluator) RunProgram(programString string) treeNodes.SmalltalkObjectInterface {
	programNode, ok := e.programCache[programString]
	if !ok {
		programNode = parser.InitializeParserFor(programString)
		e.programCache[programString] = programNode
	}
	return e.EvaluateProgramNode(programNode)
}

func (e *Evaluator) EvaluateProgramNode(programNode treeNodes.ProgramNodeInterface) treeNodes.SmalltalkObjectInterface {
	var result treeNodes.SmalltalkObjectInterface
	localScope := new(treeNodes.Scope).Initialize()
	localScope.OuterScope = e.globalScope

	if e.globalScope.IsDirty() || programNode.GetLastValue() == nil {
		result = programNode.Eval(localScope)
		programNode.SetLastValue(result)
		e.globalScope.Clean()
	} else {
		result = programNode.GetLastValue()
	}

	return result
}

func (e *Evaluator) EvaluateToString(programString string) string {
	resultObject := e.RunProgram(programString)
	return resultObject.(*treeNodes.SmalltalkString).GetValue()
}

func (e *Evaluator) EvaluateToFloat64(programString string) float64 {
	resultObject := e.RunProgram(programString)
	return resultObject.(*treeNodes.SmalltalkNumber).GetValue()
}

func (e *Evaluator) EvaluateToInt64(programString string) int64 {
	return int64(e.EvaluateToFloat64(programString))
}

func (e *Evaluator) EvaluateToBool(programString string) bool {
	resultObject := e.RunProgram(programString)
	return resultObject.(*treeNodes.SmalltalkBoolean).GetValue()
}

//scope-related delegations
func (e *Evaluator) SetVar(name string, value treeNodes.SmalltalkObjectInterface) treeNodes.SmalltalkObjectInterface {
	return e.globalScope.SetVar(name, value)
}

func (e *Evaluator) SetStringVar(name string, value string) treeNodes.SmalltalkObjectInterface {
	return e.globalScope.SetStringVar(name, value)
}

func (e *Evaluator) SetNumberVar(name string, value float64) treeNodes.SmalltalkObjectInterface {
	return e.globalScope.SetNumberVar(name, value)
}

func (e *Evaluator) SetBoolVar(name string, value bool) treeNodes.SmalltalkObjectInterface {
	return e.globalScope.SetBoolVar(name, value)
}

func (e *Evaluator) FindValueByName(name string) (treeNodes.SmalltalkObjectInterface, bool) {
	return e.globalScope.FindValueByName(name)
}
