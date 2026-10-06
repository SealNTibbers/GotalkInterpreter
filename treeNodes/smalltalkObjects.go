package treeNodes

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
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

// ErrZeroDivide is returned (wrapped) for division by zero, as Smalltalk's ZeroDivide.
var ErrZeroDivide = errors.New("ZeroDivide")

type SmalltalkObjectInterface interface {
	TypeOf() string
	Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error)
	Value() SmalltalkObjectInterface
}

// ---------------------------------------------------------------------------------------------- dispatch

type method struct {
	arity int
	fn    func(receiver SmalltalkObjectInterface, args []SmalltalkObjectInterface) (SmalltalkObjectInterface, error)
}

type methodTable map[string]method

// dispatch looks the selector up in the receiver's table, then in the messages every object understands.
func dispatch(receiver SmalltalkObjectInterface, table methodTable, selector string, args []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	m, ok := table[selector]
	if !ok {
		m, ok = objectMethods[selector]
	}
	if !ok {
		return nil, fmt.Errorf("%s does not understand #%s", className(receiver), selector)
	}
	if len(args) != m.arity {
		return nil, fmt.Errorf("#%s expects %d argument(s), got %d", selector, m.arity, len(args))
	}
	return m.fn(receiver, args)
}

func className(object SmalltalkObjectInterface) string {
	switch object.(type) {
	case *SmalltalkNumber:
		return "Number"
	case *SmalltalkString:
		return "String"
	case *SmalltalkBoolean:
		return "Boolean"
	case *SmalltalkArray:
		return "Array"
	case *SmalltalkBlock, *Deferred:
		return "BlockClosure"
	case *SmalltalkUndefinedObject, nil:
		return "UndefinedObject"
	default:
		return "Object"
	}
}

// resolve turns a variable's stored value into the value a program sees: deferred values are evaluated,
// and a missing value is nil.
func resolve(object SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	if object == nil {
		return NewSmalltalkUndefinedObject(), nil
	}
	if deferred, ok := object.(*Deferred); ok {
		value, err := deferred.SmalltalkBlock.call(nil)
		if err != nil {
			return nil, err
		}
		return resolve(value)
	}
	return object, nil
}

// valueOf is what `x value` gives where Smalltalk accepts a block or a plain object (ifTrue:, and:, ifNil:, ...).
func valueOf(object SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	if block, ok := object.(*SmalltalkBlock); ok {
		return block.call(nil)
	}
	return resolve(object)
}

// cull evaluates a block with the argument if it takes one, without it if it takes none.
func cull(object SmalltalkObjectInterface, argument SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	if block, ok := object.(*SmalltalkBlock); ok && block.numArgs() == 1 {
		return block.call([]SmalltalkObjectInterface{argument})
	}
	return valueOf(object)
}

func numberArg(selector string, object SmalltalkObjectInterface) (float64, error) {
	if number, ok := object.(*SmalltalkNumber); ok {
		return number.value, nil
	}
	return 0, fmt.Errorf("#%s expects a Number argument, got %s", selector, className(object))
}

func integerArg(selector string, object SmalltalkObjectInterface) (int64, error) {
	value, err := numberArg(selector, object)
	if err != nil {
		return 0, err
	}
	if value != math.Trunc(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("#%s expects an Integer argument, got %s", selector, formatNumber(value))
	}
	return int64(value), nil
}

func stringArg(selector string, object SmalltalkObjectInterface) (string, error) {
	if str, ok := object.(*SmalltalkString); ok {
		return str.value, nil
	}
	return "", fmt.Errorf("#%s expects a String argument, got %s", selector, className(object))
}

func booleanArg(selector string, object SmalltalkObjectInterface) (bool, error) {
	if boolean, ok := object.(*SmalltalkBoolean); ok {
		return boolean.value, nil
	}
	return false, fmt.Errorf("#%s expects a Boolean, got %s", selector, className(object))
}

func zeroDivide(selector string) error {
	return fmt.Errorf("%w in #%s", ErrZeroDivide, selector)
}

// equalObjects is Smalltalk's = : numbers, strings, booleans and arrays compare by value, other objects by identity.
func equalObjects(a, b SmalltalkObjectInterface) bool {
	switch x := a.(type) {
	case *SmalltalkNumber:
		y, ok := b.(*SmalltalkNumber)
		return ok && x.value == y.value
	case *SmalltalkString:
		y, ok := b.(*SmalltalkString)
		return ok && x.value == y.value
	case *SmalltalkBoolean:
		y, ok := b.(*SmalltalkBoolean)
		return ok && x.value == y.value
	case *SmalltalkUndefinedObject:
		_, ok := b.(*SmalltalkUndefinedObject)
		return ok
	case *SmalltalkArray:
		y, ok := b.(*SmalltalkArray)
		if !ok || len(x.array) != len(y.array) {
			return false
		}
		for i := range x.array {
			if !equalObjects(x.array[i], y.array[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// identicalObjects is == : immediate values (numbers, booleans, nil) are identical when equal.
func identicalObjects(a, b SmalltalkObjectInterface) bool {
	switch a.(type) {
	case *SmalltalkNumber, *SmalltalkBoolean, *SmalltalkUndefinedObject:
		return equalObjects(a, b)
	default:
		return a == b
	}
}

// formatNumber prints a number the way Amber (JavaScript) does, so labels match the editor:
// integers without a fraction, exponent notation below 1e-7 and from 1e21.
func formatNumber(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		return "0"
	}
	if abs := math.Abs(value); abs >= 1e-7 && abs < 1e21 {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}
	text := strconv.FormatFloat(value, 'e', -1, 64)
	mantissa, exponent, _ := strings.Cut(text, "e")
	sign := exponent[:1]
	digits := strings.TrimLeft(exponent[1:], "0")
	return mantissa + "e" + sign + digits
}

func printString(object SmalltalkObjectInterface) string {
	switch x := object.(type) {
	case *SmalltalkNumber:
		return formatNumber(x.value)
	case *SmalltalkString:
		return "'" + strings.ReplaceAll(x.value, "'", "''") + "'"
	case *SmalltalkBoolean:
		return strconv.FormatBool(x.value)
	case *SmalltalkArray:
		parts := make([]string, len(x.array))
		for i, each := range x.array {
			parts[i] = printString(each)
		}
		return "#(" + strings.Join(parts, " ") + ")"
	case *SmalltalkUndefinedObject, nil:
		return "nil"
	default:
		return "a " + className(object)
	}
}

func displayString(object SmalltalkObjectInterface) string {
	if str, ok := object.(*SmalltalkString); ok {
		return str.value
	}
	return printString(object)
}

func newNumber(value float64) (SmalltalkObjectInterface, error) {
	return NewSmalltalkNumber(value), nil
}

func newBoolean(value bool) (SmalltalkObjectInterface, error) {
	return NewSmalltalkBoolean(value), nil
}

func newString(value string) (SmalltalkObjectInterface, error) {
	return NewSmalltalkString(value), nil
}

// ---------------------------------------------------------------------------------------------- Object

func isType(check func(SmalltalkObjectInterface) bool) method {
	return method{0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return newBoolean(check(r))
	}}
}

var objectMethods methodTable

func init() {
	objectMethods = methodTable{
		`=`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(equalObjects(r, a[0]))
		}},
		`~=`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(!equalObjects(r, a[0]))
		}},
		`==`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(identicalObjects(r, a[0]))
		}},
		`~~`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(!identicalObjects(r, a[0]))
		}},
		`value`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return r, nil
		}},
		`yourself`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return r, nil
		}},
		`printString`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newString(printString(r))
		}},
		`displayString`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newString(displayString(r))
		}},
		`isNil`:     isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkUndefinedObject); return ok }),
		`notNil`:    isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkUndefinedObject); return !ok }),
		`isNumber`:  isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkNumber); return ok }),
		`isString`:  isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkString); return ok }),
		`isBoolean`: isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkBoolean); return ok }),
		`isArray`:   isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkArray); return ok }),
		`isBlock`:   isType(func(r SmalltalkObjectInterface) bool { _, ok := r.(*SmalltalkBlock); return ok }),
		`ifNil:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return r, nil
		}},
		`ifNotNil:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return cull(a[0], r)
		}},
		`ifNil:ifNotNil:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return cull(a[1], r)
		}},
		`ifNotNil:ifNil:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return cull(a[0], r)
		}},
	}
}

type SmalltalkObject struct {
}

func (obj *SmalltalkObject) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return nil, fmt.Errorf("Object does not understand #%s", name)
}

// ---------------------------------------------------------------------------------------------- nil

var undefinedMethods = methodTable{
	`ifNil:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return valueOf(a[0])
	}},
	`ifNotNil:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return r, nil
	}},
	`ifNil:ifNotNil:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return valueOf(a[0])
	}},
	`ifNotNil:ifNil:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return valueOf(a[1])
	}},
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
	return dispatch(n, undefinedMethods, name, params)
}

func (n *SmalltalkUndefinedObject) TypeOf() string {
	return UNDEFINED_OBJ
}

// ---------------------------------------------------------------------------------------------- Number

// unaryNumber defines a unary message computing a number from the receiver's value.
func unaryNumber(f func(float64) float64) method {
	return method{0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return newNumber(f(r.(*SmalltalkNumber).value))
	}}
}

// binaryNumber defines a message with one Number argument.
func binaryNumber(selector string, f func(x, y float64) (SmalltalkObjectInterface, error)) method {
	return method{1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		y, err := numberArg(selector, a[0])
		if err != nil {
			return nil, err
		}
		return f(r.(*SmalltalkNumber).value, y)
	}}
}

func arithmetic(selector string, f func(x, y float64) float64) method {
	return binaryNumber(selector, func(x, y float64) (SmalltalkObjectInterface, error) { return newNumber(f(x, y)) })
}

func divisionLike(selector string, f func(x, y float64) float64) method {
	return binaryNumber(selector, func(x, y float64) (SmalltalkObjectInterface, error) {
		if y == 0 {
			return nil, zeroDivide(selector)
		}
		return newNumber(f(x, y))
	})
}

func comparison(selector string, f func(x, y float64) bool) method {
	return binaryNumber(selector, func(x, y float64) (SmalltalkObjectInterface, error) { return newBoolean(f(x, y)) })
}

// FlooredModulo is Smalltalk's \\ : the result has the sign of the divisor (-7 \\ 2 = 1, 7.5 \\ 2 = 1.5).
func FlooredModulo(x, y float64) float64 {
	return x - math.Floor(x/y)*y
}

func sign(value float64) float64 {
	switch {
	case value > 0:
		return 1
	case value < 0:
		return -1
	default:
		return 0
	}
}

func isIntegral(value float64) bool {
	return value == math.Trunc(value) && !math.IsInf(value, 0)
}

func padLeft(text string, padding string, width int) string {
	for utf8.RuneCountInString(text) < width {
		text = padding + text
	}
	return text
}

var numberMethods methodTable

func init() {
	numberMethods = methodTable{
		`+`:           arithmetic(`+`, func(x, y float64) float64 { return x + y }),
		`-`:           arithmetic(`-`, func(x, y float64) float64 { return x - y }),
		`*`:           arithmetic(`*`, func(x, y float64) float64 { return x * y }),
		`/`:           divisionLike(`/`, func(x, y float64) float64 { return x / y }),
		`//`:          divisionLike(`//`, func(x, y float64) float64 { return math.Floor(x / y) }),
		`\\`:          divisionLike(`\\`, FlooredModulo),
		`rem:`:        divisionLike(`rem:`, func(x, y float64) float64 { return x - math.Trunc(x/y)*y }),
		`quo:`:        divisionLike(`quo:`, func(x, y float64) float64 { return math.Trunc(x / y) }),
		`roundTo:`:    divisionLike(`roundTo:`, func(x, y float64) float64 { return math.Round(x/y) * y }),
		`truncateTo:`: divisionLike(`truncateTo:`, func(x, y float64) float64 { return math.Trunc(x/y) * y }),
		`raisedTo:`: binaryNumber(`raisedTo:`, func(x, y float64) (SmalltalkObjectInterface, error) {
			result := math.Pow(x, y)
			if math.IsNaN(result) && !math.IsNaN(x) && !math.IsNaN(y) {
				return nil, fmt.Errorf("#raisedTo: cannot raise %s to %s", formatNumber(x), formatNumber(y))
			}
			return newNumber(result)
		}),
		`log:`:    binaryNumber(`log:`, func(x, y float64) (SmalltalkObjectInterface, error) { return newNumber(math.Log(x) / math.Log(y)) }),
		`arcTan:`: arithmetic(`arcTan:`, math.Atan2),
		`max:`:    arithmetic(`max:`, math.Max),
		`min:`:    arithmetic(`min:`, math.Min),
		`>`:       comparison(`>`, func(x, y float64) bool { return x > y }),
		`>=`:      comparison(`>=`, func(x, y float64) bool { return x >= y }),
		`<`:       comparison(`<`, func(x, y float64) bool { return x < y }),
		`<=`:      comparison(`<=`, func(x, y float64) bool { return x <= y }),
		`between:and:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			low, err := numberArg(`between:and:`, a[0])
			if err != nil {
				return nil, err
			}
			high, err := numberArg(`between:and:`, a[1])
			if err != nil {
				return nil, err
			}
			value := r.(*SmalltalkNumber).value
			return newBoolean(low <= value && value <= high)
		}},
		`abs`:              unaryNumber(math.Abs),
		`negated`:          unaryNumber(func(x float64) float64 { return -x }),
		`sqrt`:             unaryNumber(math.Sqrt),
		`sqr`:              unaryNumber(func(x float64) float64 { return x * x }),
		`squared`:          unaryNumber(func(x float64) float64 { return x * x }),
		`sin`:              unaryNumber(math.Sin),
		`cos`:              unaryNumber(math.Cos),
		`tan`:              unaryNumber(math.Tan),
		`arcSin`:           unaryNumber(math.Asin),
		`arcCos`:           unaryNumber(math.Acos),
		`arcTan`:           unaryNumber(math.Atan),
		`ln`:               unaryNumber(math.Log),
		`log`:              unaryNumber(math.Log10),
		`exp`:              unaryNumber(math.Exp),
		`rounded`:          unaryNumber(math.Round),
		`truncated`:        unaryNumber(math.Trunc),
		`asInteger`:        unaryNumber(math.Trunc),
		`floor`:            unaryNumber(math.Floor),
		`ceiling`:          unaryNumber(math.Ceil),
		`fractionPart`:     unaryNumber(func(x float64) float64 { return x - math.Trunc(x) }),
		`integerPart`:      unaryNumber(math.Trunc),
		`sign`:             unaryNumber(sign),
		`degreesToRadians`: unaryNumber(func(x float64) float64 { return x * math.Pi / 180.0 }),
		`radiansToDegrees`: unaryNumber(func(x float64) float64 { return x * 180.0 / math.Pi }),
		`asFloat`:          unaryNumber(func(x float64) float64 { return x }),
		`asNumber`:         unaryNumber(func(x float64) float64 { return x }),
		`isZero`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(r.(*SmalltalkNumber).value == 0)
		}},
		`even`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			value := r.(*SmalltalkNumber).value
			return newBoolean(isIntegral(value) && FlooredModulo(value, 2) == 0)
		}},
		`odd`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			value := r.(*SmalltalkNumber).value
			return newBoolean(isIntegral(value) && FlooredModulo(value, 2) == 1)
		}},
		`asString`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newString(formatNumber(r.(*SmalltalkNumber).value))
		}},
		// 255 printString: 16 = 'FF'
		`printString:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			radix, err := integerArg(`printString:`, a[0])
			if err != nil {
				return nil, err
			}
			if radix < 2 || radix > 36 {
				return nil, fmt.Errorf("#printString: radix must be between 2 and 36, got %d", radix)
			}
			value := r.(*SmalltalkNumber).value
			if !isIntegral(value) {
				return nil, fmt.Errorf("#printString: expects an Integer receiver, got %s", formatNumber(value))
			}
			return newString(strings.ToUpper(strconv.FormatInt(int64(value), int(radix))))
		}},
		// 7 printPaddedWith: $0 to: 3 = '007', -5 printPaddedWith: $0 to: 3 = '-05'
		`printPaddedWith:to:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			padding, err := stringArg(`printPaddedWith:to:`, a[0])
			if err != nil {
				return nil, err
			}
			if utf8.RuneCountInString(padding) != 1 {
				return nil, fmt.Errorf("#printPaddedWith:to: expects a Character, got '%s'", padding)
			}
			width, err := integerArg(`printPaddedWith:to:`, a[1])
			if err != nil {
				return nil, err
			}
			value := r.(*SmalltalkNumber).value
			if !isIntegral(value) {
				return nil, fmt.Errorf("#printPaddedWith:to: expects an Integer receiver, got %s", formatNumber(value))
			}
			if value < 0 {
				return newString("-" + padLeft(formatNumber(-value), padding, int(width)-1))
			}
			return newString(padLeft(formatNumber(value), padding, int(width)))
		}},
		// 3.14159 printShowingDecimalPlaces: 2 = '3.14'
		`printShowingDecimalPlaces:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			places, err := integerArg(`printShowingDecimalPlaces:`, a[0])
			if err != nil {
				return nil, err
			}
			if places < 0 {
				return nil, fmt.Errorf("#printShowingDecimalPlaces: expects a non-negative Integer, got %d", places)
			}
			return newString(strconv.FormatFloat(r.(*SmalltalkNumber).value, 'f', int(places), 64))
		}},
	}
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
	return dispatch(n, numberMethods, name, params)
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

// ---------------------------------------------------------------------------------------------- String

func unaryString(f func(string) (SmalltalkObjectInterface, error)) method {
	return method{0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return f(r.(*SmalltalkString).value)
	}}
}

func stringComparison(selector string, f func(x, y string) bool) method {
	return method{1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		y, err := stringArg(selector, a[0])
		if err != nil {
			return nil, err
		}
		return newBoolean(f(r.(*SmalltalkString).value, y))
	}}
}

// characterAt returns the 1-based character as a one-character string, or false when out of range.
func characterAt(text string, index int64) (SmalltalkObjectInterface, bool) {
	runes := []rune(text)
	if index < 1 || index > int64(len(runes)) {
		return nil, false
	}
	return NewSmalltalkString(string(runes[index-1])), true
}

var stringMethods methodTable

func init() {
	stringMethods = methodTable{
		`,`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			other, err := stringArg(`,`, a[0])
			if err != nil {
				return nil, err
			}
			return newString(r.(*SmalltalkString).value + other)
		}},
		`<`:                  stringComparison(`<`, func(x, y string) bool { return x < y }),
		`<=`:                 stringComparison(`<=`, func(x, y string) bool { return x <= y }),
		`>`:                  stringComparison(`>`, func(x, y string) bool { return x > y }),
		`>=`:                 stringComparison(`>=`, func(x, y string) bool { return x >= y }),
		`includesSubstring:`: stringComparison(`includesSubstring:`, strings.Contains),
		`size`: unaryString(func(s string) (SmalltalkObjectInterface, error) {
			return newNumber(float64(utf8.RuneCountInString(s)))
		}),
		`isEmpty`:     unaryString(func(s string) (SmalltalkObjectInterface, error) { return newBoolean(s == "") }),
		`notEmpty`:    unaryString(func(s string) (SmalltalkObjectInterface, error) { return newBoolean(s != "") }),
		`asString`:    unaryString(newString),
		`asSymbol`:    unaryString(newString),
		`asUppercase`: unaryString(func(s string) (SmalltalkObjectInterface, error) { return newString(strings.ToUpper(s)) }),
		`asLowercase`: unaryString(func(s string) (SmalltalkObjectInterface, error) { return newString(strings.ToLower(s)) }),
		`reversed`: unaryString(func(s string) (SmalltalkObjectInterface, error) {
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return newString(string(runes))
		}),
		`asNumber`: unaryString(func(s string) (SmalltalkObjectInterface, error) {
			value, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return nil, fmt.Errorf("#asNumber: '%s' is not a number", s)
			}
			return newNumber(value)
		}),
		`first`: unaryString(func(s string) (SmalltalkObjectInterface, error) {
			if character, ok := characterAt(s, 1); ok {
				return character, nil
			}
			return nil, errors.New("#first: the string is empty")
		}),
		`last`: unaryString(func(s string) (SmalltalkObjectInterface, error) {
			if character, ok := characterAt(s, int64(utf8.RuneCountInString(s))); ok {
				return character, nil
			}
			return nil, errors.New("#last: the string is empty")
		}),
		`at:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			index, err := integerArg(`at:`, a[0])
			if err != nil {
				return nil, err
			}
			if character, ok := characterAt(r.(*SmalltalkString).value, index); ok {
				return character, nil
			}
			return nil, fmt.Errorf("#at: index %d out of bounds (size %d)", index, utf8.RuneCountInString(r.(*SmalltalkString).value))
		}},
		`at:ifAbsent:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			if index, err := integerArg(`at:ifAbsent:`, a[0]); err == nil {
				if character, ok := characterAt(r.(*SmalltalkString).value, index); ok {
					return character, nil
				}
			}
			return valueOf(a[1])
		}},
		`copyFrom:to:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			from, err := integerArg(`copyFrom:to:`, a[0])
			if err != nil {
				return nil, err
			}
			to, err := integerArg(`copyFrom:to:`, a[1])
			if err != nil {
				return nil, err
			}
			runes := []rune(r.(*SmalltalkString).value)
			if to < from {
				return newString("")
			}
			if from < 1 || to > int64(len(runes)) {
				return nil, fmt.Errorf("#copyFrom:to: range %d to %d out of bounds (size %d)", from, to, len(runes))
			}
			return newString(string(runes[from-1 : to]))
		}},
	}
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

func (s *SmalltalkString) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return dispatch(s, stringMethods, name, params)
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

// ---------------------------------------------------------------------------------------------- Boolean

func booleanOperator(selector string, f func(x, y bool) bool) method {
	return method{1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		y, err := booleanArg(selector, a[0])
		if err != nil {
			return nil, err
		}
		return newBoolean(f(r.(*SmalltalkBoolean).value, y))
	}}
}

// conditional evaluates trueBranch or falseBranch (blocks or plain objects); a missing branch gives nil.
func conditional(r SmalltalkObjectInterface, trueBranch, falseBranch SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	branch := falseBranch
	if r.(*SmalltalkBoolean).value {
		branch = trueBranch
	}
	if branch == nil {
		return NewSmalltalkUndefinedObject(), nil
	}
	return valueOf(branch)
}

var booleanMethods methodTable

func init() {
	booleanMethods = methodTable{
		`&`: booleanOperator(`&`, func(x, y bool) bool { return x && y }),
		`|`: booleanOperator(`|`, func(x, y bool) bool { return x || y }),
		`xor:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			other, err := valueOf(a[0])
			if err != nil {
				return nil, err
			}
			y, err := booleanArg(`xor:`, other)
			if err != nil {
				return nil, err
			}
			return newBoolean(r.(*SmalltalkBoolean).value != y)
		}},
		`and:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			if !r.(*SmalltalkBoolean).value {
				return r, nil
			}
			return valueOf(a[0])
		}},
		`or:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			if r.(*SmalltalkBoolean).value {
				return r, nil
			}
			return valueOf(a[0])
		}},
		`not`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(!r.(*SmalltalkBoolean).value)
		}},
		`asString`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newString(printString(r))
		}},
		`ifTrue:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return conditional(r, a[0], nil)
		}},
		`ifFalse:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return conditional(r, nil, a[0])
		}},
		`ifTrue:ifFalse:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return conditional(r, a[0], a[1])
		}},
		`ifFalse:ifTrue:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return conditional(r, a[1], a[0])
		}},
	}
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
	return dispatch(b, booleanMethods, name, params)
}

// ---------------------------------------------------------------------------------------------- Block

func blockValue(arity int) method {
	return method{arity, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		return r.(*SmalltalkBlock).call(a)
	}}
}

var blockMethods methodTable

func init() {
	blockMethods = methodTable{
		`value`:                    blockValue(0),
		`value:`:                   blockValue(1),
		`value:value:`:             blockValue(2),
		`value:value:value:`:       blockValue(3),
		`value:value:value:value:`: blockValue(4),
		`valueWithArguments:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			arguments, ok := a[0].(*SmalltalkArray)
			if !ok {
				return nil, fmt.Errorf("#valueWithArguments: expects an Array, got %s", className(a[0]))
			}
			return r.(*SmalltalkBlock).call(arguments.array)
		}},
		`numArgs`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newNumber(float64(r.(*SmalltalkBlock).numArgs()))
		}},
	}
}

type SmalltalkBlock struct {
	*SmalltalkObject
	block *BlockNode
	scope *Scope
}

func (b *SmalltalkBlock) numArgs() int {
	return len(b.block.arguments)
}

// call evaluates the block body in a new scope holding its arguments and temporaries.
func (b *SmalltalkBlock) call(args []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	if len(args) != b.numArgs() {
		return nil, fmt.Errorf("wrong number of arguments: the block takes %d, got %d", b.numArgs(), len(args))
	}
	scope := newChildScope(b.scope)
	for i, argument := range b.block.arguments {
		scope.SetVar(argument.GetName(), args[i])
	}
	if b.block.body == nil {
		return NewSmalltalkUndefinedObject(), nil
	}
	return b.block.body.Eval(scope)
}

// Value evaluates a block without arguments. It returns nil on errors; use `value` through Perform to get them.
func (b *SmalltalkBlock) Value() SmalltalkObjectInterface {
	result, err := b.call(nil)
	if err != nil {
		return NewSmalltalkUndefinedObject()
	}
	return result
}

func (b *SmalltalkBlock) TypeOf() string {
	return BLOCK_OBJ
}

func (b *SmalltalkBlock) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	return dispatch(b, blockMethods, name, params)
}

// ---------------------------------------------------------------------------------------------- Array

// elementwise applies a Number operation to each element, with a Number or an Array of the same size.
func elementwise(selector string) method {
	return method{1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
		receiver := r.(*SmalltalkArray)
		other, isArray := a[0].(*SmalltalkArray)
		if isArray && len(other.array) != len(receiver.array) {
			return nil, fmt.Errorf("#%s: arrays have different sizes (%d and %d)", selector, len(receiver.array), len(other.array))
		}
		result := &SmalltalkArray{array: make([]SmalltalkObjectInterface, len(receiver.array))}
		for i, each := range receiver.array {
			if _, ok := each.(*SmalltalkNumber); !ok {
				return nil, fmt.Errorf("#%s: array element %d is a %s, not a Number", selector, i+1, className(each))
			}
			argument := a[0]
			if isArray {
				argument = other.array[i]
			}
			value, err := each.Perform(selector, []SmalltalkObjectInterface{argument})
			if err != nil {
				return nil, err
			}
			result.array[i] = value
		}
		return result, nil
	}}
}

func (a *SmalltalkArray) elementAt(index int64) (SmalltalkObjectInterface, bool) {
	if index < 1 || index > int64(len(a.array)) {
		return nil, false
	}
	return a.array[index-1], true
}

var arrayMethods methodTable

func init() {
	arrayMethods = methodTable{
		`at:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			index, err := integerArg(`at:`, a[0])
			if err != nil {
				return nil, err
			}
			if element, ok := r.(*SmalltalkArray).elementAt(index); ok {
				return element, nil
			}
			return nil, fmt.Errorf("#at: index %d out of bounds (size %d)", index, len(r.(*SmalltalkArray).array))
		}},
		`at:ifAbsent:`: {2, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			if index, err := integerArg(`at:ifAbsent:`, a[0]); err == nil {
				if element, ok := r.(*SmalltalkArray).elementAt(index); ok {
					return element, nil
				}
			}
			return valueOf(a[1])
		}},
		`size`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newNumber(float64(len(r.(*SmalltalkArray).array)))
		}},
		`isEmpty`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(len(r.(*SmalltalkArray).array) == 0)
		}},
		`notEmpty`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			return newBoolean(len(r.(*SmalltalkArray).array) != 0)
		}},
		`first`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			if element, ok := r.(*SmalltalkArray).elementAt(1); ok {
				return element, nil
			}
			return nil, errors.New("#first: the array is empty")
		}},
		`last`: {0, func(r SmalltalkObjectInterface, _ []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			array := r.(*SmalltalkArray)
			if element, ok := array.elementAt(int64(len(array.array))); ok {
				return element, nil
			}
			return nil, errors.New("#last: the array is empty")
		}},
		`includes:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			for _, each := range r.(*SmalltalkArray).array {
				if equalObjects(each, a[0]) {
					return newBoolean(true)
				}
			}
			return newBoolean(false)
		}},
		`indexOf:`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			for i, each := range r.(*SmalltalkArray).array {
				if equalObjects(each, a[0]) {
					return newNumber(float64(i + 1))
				}
			}
			return newNumber(0)
		}},
		`,`: {1, func(r SmalltalkObjectInterface, a []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
			other, ok := a[0].(*SmalltalkArray)
			if !ok {
				return nil, fmt.Errorf("#, expects an Array argument, got %s", className(a[0]))
			}
			result := &SmalltalkArray{}
			result.array = append(append(result.array, r.(*SmalltalkArray).array...), other.array...)
			return result, nil
		}},
		`+`:  elementwise(`+`),
		`-`:  elementwise(`-`),
		`*`:  elementwise(`*`),
		`/`:  elementwise(`/`),
		`\\`: elementwise(`\\`),
		`//`: elementwise(`//`),
	}
}

type SmalltalkArray struct {
	SmalltalkObject
	array []SmalltalkObjectInterface
}

func NewSmalltalkArray(elements []SmalltalkObjectInterface) *SmalltalkArray {
	return &SmalltalkArray{array: elements}
}

// GetValueAt returns the element at a 0-based index (Go side), or nil when out of range.
func (a *SmalltalkArray) GetValueAt(index int64) SmalltalkObjectInterface {
	if index < 0 || index >= int64(len(a.array)) {
		return nil
	}
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
		case UNDEFINED_OBJ:
			interfaceSlice[i] = nil
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
	return dispatch(a, arrayMethods, name, params)
}

// ---------------------------------------------------------------------------------------------- Deferred

// Deferred is a variable value computed by a block each time it is read.
type Deferred struct {
	*SmalltalkBlock
}

func (d *Deferred) TypeOf() string {
	return DEFERRED
}

func (d *Deferred) Value() SmalltalkObjectInterface {
	value, err := resolve(d)
	if err != nil {
		return NewSmalltalkUndefinedObject()
	}
	return value
}

func (d *Deferred) Perform(name string, params []SmalltalkObjectInterface) (SmalltalkObjectInterface, error) {
	value, err := resolve(d)
	if err != nil {
		return nil, err
	}
	return value.Perform(name, params)
}

func NewDeferred(blockNode *BlockNode, scope *Scope) *Deferred {
	return &Deferred{&SmalltalkBlock{&SmalltalkObject{}, blockNode, scope}}
}
