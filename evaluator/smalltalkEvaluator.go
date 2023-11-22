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
	programNode, _ := parser.InitializeParserFor(codeString)
	return evaluator.EvaluateProgram(programNode)
}

func TestEvalWithScope(codeString string, scope *treeNodes.Scope) treeNodes.SmalltalkObjectInterface {
	evaluator := NewEvaluatorWithGlobalScope(scope)
	programNode, _ := parser.InitializeParserFor(codeString)
	return evaluator.EvaluateProgram(programNode)
}

//real world API
func NewSmalltalkVM() *Evaluator {
	globalScope := new(treeNodes.Scope).Initialize()
	return NewEvaluatorWithGlobalScope(globalScope)
}

func NewSmalltalkWorkspace() *Evaluator {
	globalScope := new(treeNodes.Scope).Initialize()
	evaluator := NewEvaluatorWithGlobalScope(globalScope)
	evaluator.workspaceScope = new(treeNodes.Scope).Initialize()
	return evaluator
}

func NewEvaluatorWithGlobalScope(global *treeNodes.Scope) *Evaluator {
	evaluator := new(Evaluator)
	evaluator.programCache = make(map[string]treeNodes.ProgramNodeInterface)
	evaluator.globalScope = global
	return evaluator
}

type Evaluator struct {
	globalScope    *treeNodes.Scope
	programCache   map[string]treeNodes.ProgramNodeInterface
	workspaceScope *treeNodes.Scope
}

func (e *Evaluator) SetGlobalScope(scope *treeNodes.Scope) *Evaluator {
	e.globalScope = scope
	return e
}

func (e *Evaluator) GetGlobalScope() *treeNodes.Scope {
	return e.globalScope
}

func (e *Evaluator) RunProgram(programString string) treeNodes.SmalltalkObjectInterface {
	_, ok := e.programCache[programString]
	if !ok {
		initializedParser, err := parser.InitializeParserFor(programString)
		if err != nil {
			return treeNodes.NewSmalltalkString(err.Error())
		}
		e.programCache[programString] = initializedParser
	}
	evaluatorProgram := e.programCache[programString]
	return e.EvaluateProgram(evaluatorProgram)
}

func (e *Evaluator) EvaluateProgram(program treeNodes.ProgramNodeInterface) treeNodes.SmalltalkObjectInterface {
	var result treeNodes.SmalltalkObjectInterface
	var localScope *treeNodes.Scope
	if e.workspaceScope != nil {
		localScope = e.workspaceScope
	} else {
		localScope = new(treeNodes.Scope).Initialize()
	}
	localScope.OuterScope = e.globalScope

	if program.GetLastValue() == nil {
		result = program.Eval(localScope)
		program.SetLastValue(result)
	} else {
		result = program.GetLastValue()
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

func (e *Evaluator) EvaluateToInterface(programString string) interface{} {
	resultObject := e.RunProgram(programString)
	if resultObject == nil {
		return nil
	}
	switch resultObject.TypeOf() {
	case treeNodes.NUMBER_OBJ:
		return resultObject.(*treeNodes.SmalltalkNumber).GetValue()
	case treeNodes.STRING_OBJ:
		return resultObject.(*treeNodes.SmalltalkString).GetValue()
	case treeNodes.BOOLEAN_OBJ:
		return resultObject.(*treeNodes.SmalltalkBoolean).GetValue()
	case treeNodes.ARRAY_OBJ:
		array, err := resultObject.(*treeNodes.SmalltalkArray).GetValue()
		if err != nil {
			return nil
		}
		return array
	default:
		return nil
	}
}

func (e *Evaluator) updateCache(variableName string) {
	for _, evaluatorProgram := range e.programCache {
		needsUpdate := false
		for _, variable := range evaluatorProgram.GetVariables() {
			if variable == variableName {
				needsUpdate = true
			}
		}
		if needsUpdate {
			evaluatorProgram.SetLastValue(nil)
		}
	}
}

//scope-related delegations
func (e *Evaluator) SetVar(name string, value treeNodes.SmalltalkObjectInterface) treeNodes.SmalltalkObjectInterface {
	e.updateCache(name)
	return e.globalScope.SetVar(name, value)
}

func (e *Evaluator) SetStringVar(name string, value string) treeNodes.SmalltalkObjectInterface {
	e.updateCache(name)
	return e.globalScope.SetStringVar(name, value)
}

func (e *Evaluator) SetNumberVar(name string, value float64) treeNodes.SmalltalkObjectInterface {
	e.updateCache(name)
	return e.globalScope.SetNumberVar(name, value)
}

func (e *Evaluator) SetBoolVar(name string, value bool) treeNodes.SmalltalkObjectInterface {
	e.updateCache(name)
	return e.globalScope.SetBoolVar(name, value)
}

func (e *Evaluator) FindValueByName(name string) (treeNodes.SmalltalkObjectInterface, bool) {
	e.updateCache(name)
	return e.globalScope.FindValueByName(name)
}
