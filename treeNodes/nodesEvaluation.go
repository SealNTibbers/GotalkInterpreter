package treeNodes

import (
	"fmt"
	"strconv"

	"github.com/SealNTibbers/GotalkInterpreter/scanner"
)

type Scope struct {
	variables  map[string]SmalltalkObjectInterface
	OuterScope *Scope
	global     bool
}

func (s *Scope) Initialize() *Scope {
	s.variables = make(map[string]SmalltalkObjectInterface)
	return s
}

// MarkAsGlobal makes the scope hold the externally set variables (parameters). Assignments inside programs
// never write into a global scope; they shadow its variables in a local scope instead.
func (s *Scope) MarkAsGlobal() *Scope {
	s.global = true
	return s
}

func (s *Scope) SetVar(name string, value SmalltalkObjectInterface) SmalltalkObjectInterface {
	s.variables[name] = value
	return value
}

func (s *Scope) SetStringVar(name string, value string) *SmalltalkString {
	smValue := NewSmalltalkString(value)
	return s.SetVar(name, smValue).(*SmalltalkString)
}

func (s *Scope) SetNumberVar(name string, value float64) *SmalltalkNumber {
	smValue := NewSmalltalkNumber(value)
	return s.SetVar(name, smValue).(*SmalltalkNumber)
}

func (s *Scope) SetBoolVar(name string, value bool) *SmalltalkBoolean {
	smValue := NewSmalltalkBoolean(value)
	return s.SetVar(name, smValue).(*SmalltalkBoolean)
}

func (s *Scope) FindValueByName(name string) (SmalltalkObjectInterface, bool) {
	value, ok := s.variables[name]
	return value, ok
}

func (s *Scope) GetVarValue(name string) (SmalltalkObjectInterface, error) {
	for scope := s; scope != nil; scope = scope.OuterScope {
		if value, ok := scope.variables[name]; ok {
			return value, nil
		}
	}
	return nil, fmt.Errorf("undefined variable: %s", name)
}

// assign writes to the innermost local scope that defines name. If none does, the variable is created in the
// outermost local scope (the program's, or the workspace's).
func (s *Scope) assign(name string, value SmalltalkObjectInterface) {
	outermost := s
	for scope := s; scope != nil && !scope.global; scope = scope.OuterScope {
		if _, ok := scope.variables[name]; ok {
			scope.variables[name] = value
			return
		}
		outermost = scope
	}
	outermost.variables[name] = value
}

func newChildScope(outer *Scope) *Scope {
	scope := new(Scope).Initialize()
	scope.OuterScope = outer
	return scope
}

func (message *MessageNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	receiver, err := message.receiver.Eval(scope)
	if err != nil {
		return nil, err
	}
	return message.sendTo(receiver, scope)
}

// sendTo evaluates the arguments and sends the message to an already evaluated receiver.
func (message *MessageNode) sendTo(receiver SmalltalkObjectInterface, scope *Scope) (SmalltalkObjectInterface, error) {
	argObjects := make([]SmalltalkObjectInterface, len(message.arguments))
	for i, each := range message.arguments {
		argument, err := each.Eval(scope)
		if err != nil {
			return nil, err
		}
		argObjects[i] = argument
	}
	return receiver.Perform(message.GetSelector(), argObjects)
}

func (cascade *CascadeNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	receiver, err := cascade.GetReceiver().Eval(scope)
	if err != nil {
		return nil, err
	}
	var result SmalltalkObjectInterface
	for _, message := range cascade.messages {
		result, err = message.sendTo(receiver, scope)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (block *BlockNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	return &SmalltalkBlock{&SmalltalkObject{}, block, scope}, nil
}

func (sequence *SequenceNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	nilObject := NewSmalltalkUndefinedObject()
	for _, temporary := range sequence.temporaries {
		scope.SetVar(temporary.GetName(), nilObject)
	}
	var result SmalltalkObjectInterface = nilObject
	for _, each := range sequence.statements {
		value, err := each.Eval(scope)
		if err != nil {
			return nil, err
		}
		result = value
	}
	return result, nil
}

func (assignment *AssignmentNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	value, err := assignment.value.Eval(scope)
	if err != nil {
		return nil, err
	}
	scope.assign(assignment.variable.GetName(), value)
	return value, nil
}

func (variable *VariableNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	smalltalkValue, err := scope.GetVarValue(variable.GetName())
	if err != nil {
		return nil, err
	}
	return resolve(smalltalkValue)
}

func (array *LiteralArrayNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	arr := new(SmalltalkArray)
	for _, each := range array.contents {
		value, err := each.Eval(scope)
		if err != nil {
			return nil, err
		}
		arr.array = append(arr.array, value)
	}
	return arr, nil
}

func (literalValue *LiteralValueNode) Eval(scope *Scope) (SmalltalkObjectInterface, error) {
	switch typeOfLiteral := literalValue.GetTypeOfToken(); typeOfLiteral {
	case scanner.NUMBER:
		number, err := strconv.ParseFloat(literalValue.GetValue(), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number literal: %s", literalValue.GetValue())
		}
		return NewSmalltalkNumber(number), nil
	case scanner.STRING, scanner.SYMBOL, scanner.CHAR:
		return NewSmalltalkString(literalValue.GetValue()), nil
	case scanner.BOOLEAN:
		return NewSmalltalkBoolean(literalValue.GetValue() == "true"), nil
	case scanner.NIL:
		return NewSmalltalkUndefinedObject(), nil
	default:
		return nil, fmt.Errorf("unsupported literal: %s", literalValue.GetValue())
	}
}
