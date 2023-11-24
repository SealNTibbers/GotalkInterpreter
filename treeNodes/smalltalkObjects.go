package treeNodes

import (
	"errors"
	"math"
	"reflect"
)

const (
	NUMBER_OBJ    = "NUMBER"
	BOOLEAN_OBJ   = "BOOLEAN"
	STRING_OBJ    = "STRING"
	BLOCK_OBJ     = "BLOCK"
	DEFERRED      = "DEFERRED"
	ARRAY_OBJ     = "ARRAY"
	UNDEFINED_OBJ = "UNDEFINED"
)

var numberMessages = map[string]interface{}{
	`value`:            value,
	`=`:                equal,
	`~=`:               notEqual,
	`>`:                greater,
	`>=`:               greaterEqual,
	`<`:                lesser,
	`<=`:               lesserEqual,
	`+`:                plus,
	`-`:                minus,
	`*`:                mul,
	`/`:                div,
	`\\`:               mod,
	`//`:               intDiv,
	`rem:`:             rem,
	`max:`:             max,
	`min:`:             min,
	`abs`:              abs,
	`sqrt`:             sqrt,
	`sqr`:              sqr,
	`sin`:              sin,
	`cos`:              cos,
	`tan`:              tan,
	`arcSin`:           arcSin,
	`arcCos`:           arcCos,
	`arcTan`:           arcTan,
	`rounded`:          rounded,
	`truncated`:        truncated,
	`fractionPart`:     fractionPart,
	`floor`:            floor,
	`ceiling`:          ceiling,
	`negated`:          negated,
	`degreesToRadians`: degreesToRadians,
}

var booleanMessages = map[string]interface{}{
	`value`:           value,
	`=`:               boolEqual,
	`~=`:              boolNotEqual,
	`ifTrue:`:         ifTrue,
	`ifFalse:`:        ifFalse,
	`ifTrue:ifFalse:`: ifTrueIfFalse,
	`ifFalse:ifTrue:`: ifFalseIfTrue,
	`and:`:            and,
	`&`:               ampersand,
	`or:`:             or,
	`|`:               verticalBar,
	`xor:`:            xor,
	`not`:             not,
}

var blockMessages = map[string]interface{}{
	`value`:  value,
	`value:`: valueWith,
}

var arrayMessages = map[string]interface{}{
	`at:`: ValueAt,
	`+`:   arrPlus,
	`-`:   arrMinus,
	`*`:   arrMul,
	`/`:   arrDiv,
	`\\`:  arrMod,
	`//`:  arrIntDiv,
}

func value(receiver SmalltalkObjectInterface) SmalltalkObjectInterface {
	return receiver.Value()
}

func valueWith(receiver *SmalltalkBlock, arg SmalltalkObjectInterface) SmalltalkObjectInterface {
	scope := new(Scope).Initialize()
	scope.OuterScope = receiver.scope
	scope.SetVar(receiver.block.arguments[0].GetName(), arg)
	return receiver.block.body.Eval(scope)
}

func equal(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() == arg.GetValue())
}

func notEqual(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() != arg.GetValue())
}

func greater(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() > arg.GetValue())
}

func greaterEqual(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() >= arg.GetValue())
}

func lesser(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() < arg.GetValue())
}

func lesserEqual(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() <= arg.GetValue())
}

func plus(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.GetValue() + arg.GetValue())
}

func minus(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.GetValue() - arg.GetValue())
}

func mul(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.GetValue() * arg.GetValue())
}

func div(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.GetValue() / arg.GetValue())
}

func mod(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(float64(int64(receiver.GetValue()) % int64(arg.GetValue())))
}

func intDiv(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Floor(receiver.GetValue() / arg.GetValue()))
}

func rem(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	quo := math.Trunc(receiver.value / arg.value)
	return new(SmalltalkNumber).SetValue(receiver.value - (quo * arg.value))
}

func max(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	if receiver.value > arg.value {
		return receiver
	} else {
		return arg
	}
}

func min(receiver *SmalltalkNumber, arg *SmalltalkNumber) *SmalltalkNumber {
	if receiver.value > arg.value {
		return arg
	} else {
		return receiver
	}
}

func abs(receiver *SmalltalkNumber) *SmalltalkNumber {
	if receiver.value < 0 {
		return new(SmalltalkNumber).SetValue(receiver.value * -1)
	} else {
		return receiver
	}
}

func sqrt(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Sqrt(receiver.value))
}

func sqr(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Pow(receiver.value, 2))
}

func sin(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Sin(receiver.value))
}

func cos(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Cos(receiver.value))
}

func tan(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Tan(receiver.value))
}

func arcSin(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Asin(receiver.value))
}

func arcCos(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Acos(receiver.value))
}

func arcTan(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Atan(receiver.value))
}

func rounded(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Round(receiver.value))
}

func truncated(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Trunc(receiver.value))
}

func floor(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Floor(receiver.value))
}

func ceiling(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(math.Ceil(receiver.value))
}

func fractionPart(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.value - math.Trunc(receiver.value))
}

func negated(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.value * -1)
}

func degreesToRadians(receiver *SmalltalkNumber) *SmalltalkNumber {
	return new(SmalltalkNumber).SetValue(receiver.value * math.Pi / 180.0)
}

//Boolean receiver messages section
func boolEqual(receiver *SmalltalkBoolean, arg *SmalltalkBoolean) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() == arg.GetValue())
}

func boolNotEqual(receiver *SmalltalkBoolean, arg *SmalltalkBoolean) *SmalltalkBoolean {
	return new(SmalltalkBoolean).SetValue(receiver.GetValue() != arg.GetValue())
}

func and(receiver *SmalltalkBoolean, arg *SmalltalkBlock) *SmalltalkBoolean {
	if receiver.GetValue() {
		return arg.Value().(*SmalltalkBoolean)
	} else {
		return receiver
	}
}

func ampersand(receiver *SmalltalkBoolean, arg *SmalltalkBoolean) *SmalltalkBoolean {
	if receiver.GetValue() {
		return arg
	} else {
		return receiver
	}
}

func or(receiver *SmalltalkBoolean, arg *SmalltalkBlock) *SmalltalkBoolean {
	if receiver.GetValue() {
		return receiver
	} else {
		return arg.Value().(*SmalltalkBoolean)
	}
}

func verticalBar(receiver *SmalltalkBoolean, arg *SmalltalkBoolean) *SmalltalkBoolean {
	if receiver.GetValue() {
		return receiver
	} else {
		return arg
	}
}

func xor(receiver *SmalltalkBoolean, arg *SmalltalkBoolean) *SmalltalkBoolean {
	xor := !(receiver.GetValue() == arg.GetValue())
	return new(SmalltalkBoolean).SetValue(xor)
}

func not(receiver *SmalltalkBoolean) *SmalltalkBoolean {
	if receiver.GetValue() {
		return new(SmalltalkBoolean).SetValue(false)
	} else {
		return new(SmalltalkBoolean).SetValue(true)
	}
}

func ifTrue(receiver *SmalltalkBoolean, arg SmalltalkObjectInterface) SmalltalkObjectInterface {
	if receiver.GetValue() {
		return arg.Value()
	} else {
		return NewSmalltalkUndefinedObject()
	}
}

func ifFalse(receiver *SmalltalkBoolean, arg SmalltalkObjectInterface) SmalltalkObjectInterface {
	if receiver.GetValue() {
		return NewSmalltalkUndefinedObject()
	} else {
		return arg.Value()
	}
}

func ifTrueIfFalse(receiver *SmalltalkBoolean, argTrue SmalltalkObjectInterface, argFalse SmalltalkObjectInterface) SmalltalkObjectInterface {
	if receiver.GetValue() {
		return argTrue.Value()
	} else {
		return argFalse.Value()
	}
}

func ifFalseIfTrue(receiver *SmalltalkBoolean, argFalse SmalltalkObjectInterface, argTrue SmalltalkObjectInterface) SmalltalkObjectInterface {
	if receiver.GetValue() {
		return argTrue.Value()
	} else {
		return argFalse.Value()
	}
}

// Array methods
func ValueAt(receiver *SmalltalkArray, index *SmalltalkNumber) SmalltalkObjectInterface {
	return receiver.array[int64(index.value)-1]
}

func arrPlus(receiver *SmalltalkArray, number *SmalltalkNumber) SmalltalkObjectInterface {
	result := new(SmalltalkArray)
	for _, each := range receiver.array {
		result.array = append(result.array, plus(each.(*SmalltalkNumber), number))
	}

	return result
}

func arrMinus(receiver *SmalltalkArray, number *SmalltalkNumber) SmalltalkObjectInterface {
	result := new(SmalltalkArray)
	for _, each := range receiver.array {
		result.array = append(result.array, minus(each.(*SmalltalkNumber), number))
	}

	return result
}

func arrMul(receiver *SmalltalkArray, number *SmalltalkNumber) SmalltalkObjectInterface {
	result := new(SmalltalkArray)
	for _, each := range receiver.array {
		result.array = append(result.array, mul(each.(*SmalltalkNumber), number))
	}

	return result
}

func arrDiv(receiver *SmalltalkArray, number *SmalltalkNumber) SmalltalkObjectInterface {
	result := new(SmalltalkArray)
	for _, each := range receiver.array {
		result.array = append(result.array, div(each.(*SmalltalkNumber), number))
	}

	return result
}

func arrMod(receiver *SmalltalkArray, number *SmalltalkNumber) SmalltalkObjectInterface {
	result := new(SmalltalkArray)
	for _, each := range receiver.array {
		result.array = append(result.array, mod(each.(*SmalltalkNumber), number))
	}

	return result
}

func arrIntDiv(receiver *SmalltalkArray, number *SmalltalkNumber) SmalltalkObjectInterface {
	result := new(SmalltalkArray)
	for _, each := range receiver.array {
		result.array = append(result.array, intDiv(each.(*SmalltalkNumber), number))
	}

	return result
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func Call(receiver SmalltalkObjectInterface, m map[string]interface{}, name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	var receiverAndArgs []SmalltalkObjectInterface
	if receiver.TypeOf() == DEFERRED {
		receiverAndArgs = append(receiverAndArgs, receiver.Value())
	} else {
		receiverAndArgs = append(receiverAndArgs, receiver)
	}
	for _, each := range params {
		if each.TypeOf() == DEFERRED {
			receiverAndArgs = append(receiverAndArgs, each.Value())
		} else {
			receiverAndArgs = append(receiverAndArgs, each)
		}
	}
	f, ok := m[name]
	if !ok {
		err := errors.New("does not understand: " + name)
		return nil, err
	}
	function := reflect.ValueOf(f)
	if len(receiverAndArgs) != function.Type().NumIn() {
		err := errors.New("wrong parameters length")
		return nil, err
	}
	in := make([]reflect.Value, len(receiverAndArgs))
	for k, param := range receiverAndArgs {
		in[k] = reflect.ValueOf(param)
	}
	result := function.Call(in)
	return result[0].Interface().(SmalltalkObjectInterface), nil
}

type SmalltalkObjectInterface interface {
	TypeOf() string
	Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error)
	Value() SmalltalkObjectInterface
}

type SmalltalkObject struct {
}

func (obj *SmalltalkObject) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return nil, nil
}

type SmalltalkUndefinedObject struct {
	*SmalltalkObject
}

func NewSmalltalkUndefinedObject() *SmalltalkUndefinedObject {
	return &SmalltalkUndefinedObject{&SmalltalkObject{}}
}

func (n *SmalltalkUndefinedObject) Value() SmalltalkObjectInterface {
	return n
}

func (n *SmalltalkUndefinedObject) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return nil, errors.New("doesNotUnderstand")
}

func (n *SmalltalkUndefinedObject) TypeOf() string {
	return UNDEFINED_OBJ
}

type SmalltalkNumber struct {
	*SmalltalkObject
	value float64
}

func NewSmalltalkNumber(value float64) *SmalltalkNumber {
	return &SmalltalkNumber{&SmalltalkObject{}, value}
}

func (n *SmalltalkNumber) Value() SmalltalkObjectInterface {
	return n
}

func (n *SmalltalkNumber) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return Call(n, numberMessages, name, params)
}

func (n *SmalltalkNumber) TypeOf() string {
	return NUMBER_OBJ
}

func (n *SmalltalkNumber) GetValue() float64 {
	return n.value
}

func (n *SmalltalkNumber) SetValue(val float64) *SmalltalkNumber {
	n.value = val
	return n
}

type SmalltalkString struct {
	*SmalltalkObject
	value string
}

func NewSmalltalkString(value string) *SmalltalkString {
	return &SmalltalkString{&SmalltalkObject{}, value}
}

func (s *SmalltalkString) Value() SmalltalkObjectInterface {
	return s
}

func (s *SmalltalkString) TypeOf() string {
	return STRING_OBJ
}

func (s *SmalltalkString) GetValue() string {
	return s.value
}

func (s *SmalltalkString) SetValue(val string) *SmalltalkString {
	s.value = val
	return s
}

type SmalltalkBoolean struct {
	*SmalltalkObject
	value bool
}

func NewSmalltalkBoolean(value bool) *SmalltalkBoolean {
	return &SmalltalkBoolean{&SmalltalkObject{}, value}
}

func (b *SmalltalkBoolean) Value() SmalltalkObjectInterface {
	return b
}

func (b *SmalltalkBoolean) TypeOf() string {
	return BOOLEAN_OBJ
}

func (b *SmalltalkBoolean) GetValue() bool {
	return b.value
}

func (b *SmalltalkBoolean) SetValue(val bool) *SmalltalkBoolean {
	b.value = val
	return b
}

func (b *SmalltalkBoolean) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return Call(b, booleanMessages, name, params)
}

type SmalltalkBlock struct {
	*SmalltalkObject
	block *BlockNode
	scope *Scope
}

func (b *SmalltalkBlock) Value() SmalltalkObjectInterface {
	return b.block.body.Eval(b.scope)
}

func (b *SmalltalkBlock) TypeOf() string {
	return BLOCK_OBJ
}

func (b *SmalltalkBlock) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return Call(b, blockMessages, name, params)
}

type SmalltalkArray struct {
	SmalltalkObject
	array []SmalltalkObjectInterface
}

func (a *SmalltalkArray) GetValueAt(index int64) SmalltalkObjectInterface {
	return a.array[index]
}

func (a *SmalltalkArray) GetValue() ([]interface{}, error) {
	var interfaceSlice = make([]interface{}, len(a.array))
	for i, each := range a.array {
		switch each.TypeOf() {
		case NUMBER_OBJ:
			interfaceSlice[i] = each.(*SmalltalkNumber).GetValue()
		case STRING_OBJ:
			interfaceSlice[i] = each.(*SmalltalkString).GetValue()
		case BOOLEAN_OBJ:
			interfaceSlice[i] = each.(*SmalltalkBoolean).GetValue()
		case ARRAY_OBJ:
			innerArray, err := each.(*SmalltalkArray).GetValue()
			if err != nil {
				return nil, err
			}
			interfaceSlice[i] = innerArray
		default:
			return nil, errors.New(`we do not support this type "` + each.TypeOf() + `" in array`)
		}
	}
	return interfaceSlice, nil
}

func (a *SmalltalkArray) Value() SmalltalkObjectInterface {
	return a
}

func (a *SmalltalkArray) TypeOf() string {
	return ARRAY_OBJ
}

func (a *SmalltalkArray) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return Call(a, arrayMessages, name, params)
}

type Deferred struct {
	*SmalltalkBlock
}

func (d *Deferred) TypeOf() string {
	return DEFERRED
}

func NewDeferred(blockNode *BlockNode, scope *Scope) *Deferred {
	return &Deferred{&SmalltalkBlock{&SmalltalkObject{}, blockNode, scope}}
}
