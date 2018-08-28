# GotalkInterpreter

It's a simplistic Smalltalk code interpreter library written in Golang by Alex and Michael.


#### Who can use it

The entire purpose of this library is to use Smalltalk for dynamic code (string) evaluation in Golang applications. It is optimized to reevaluate same code lines with different scope (variables). Typical use case: our app read xml file with a markup and a Smalltalk code, evaluate this code and use the result. We use it to build and animate an OpenGL UI for our embedded software.

#### Why Smalltalk

Smalltalk is beautiful dynamic language with a concise and readable syntax. Also we are smalltalkers so that's why.

#### Contents

Scanner and Parser are essentially standard Smalltalk Scanner and Parser rewritten in Go. We used [VisualWorks](http://www.cincomsmalltalk.com) and [Pharo](https://pharo.org/) realizations as reference implementations.
Evaluator is the API entry point. You can expand functionality by modifying smalltalkObjects.go file.

#### Bugs, Tests etc

There are a lot of tests for everything here so we are pretty sure that this library is actually useable. You can use its tests (specifically smalltalkEvaluator_test.go) to better understand what you can do with this library.
We can rewrite some parts later just to make our Go code better and somehow expand overall functionality.

## Installation

GotalkInterpreter does not use any third party libraries. For getting it run on your machine, you just:
```go
go get github.com/SealNTibbers/GotalkInterpreter
```

## API and Examples

Result of our Smalltalk code evaluation can be number (float64 or int. Internally it's always float64), bool or string.

In our little Smalltalk we have supported limited amount of messages that is enough for our internal project but it's easily expandable.

Numbers can receive following messages:
```go
`value`           
`=`               
`~=`            
`>`             
`>=`             
`<`              
`<=`             
`+`             
`-`               
`*`             
`/`             
`\\`             
`//`             
`rem:`           
`max:`            
`min:`            
`abs`           
`sqrt`            
`sqr`             
`sin`             
`cos`             
`tan`             
`arcSin`          
`arcCos`         
`arcTan`          
`rounded`         
`truncated`       
`fractionPart`     
`floor`           
`ceiling`          
`negated`       
`degreesToRadians`
```

Booleans can receive following messages:
```go
`value`
`=`
`~=`
`ifTrue:`
`ifFalse:`
`ifTrue:ifFalse:`
`ifFalse:ifTrue:`
`and:`
`&`
`or:`
`|`
`xor:`
`not`
```

Blocks can receive following messages:
```go
`value`
`value:`
```

Arrays can receive following messages:
```go
`at:`
`+`
`-`
`*`
`/`
`\\`
`//`
```

#### Low-lewel API example
```go
//so we have smalltalk code string and want to evaluate it `angle\\10/10-0.9*10`

globalScope := new(treeNodes.Scope).Initialize()
globalScope.SetVar("angle", treeNodes.NewSmalltalkNumber(25))

evaluator = evaluator.NewEvaluatorWithGlobalScope(globalScope)
resultObject = evaluator.RunProgram(`angle\\10/10-0.9*10`)

result = resultObject.(*treeNodes.SmalltalkNumber).GetValue()
//so now we have result which is float64 and equals to -4
```

#### Higher level API example. See TestAPI func in smalltalkEvaluator_test.go
```go
vm := NewSmalltalkVM()
vm.SetNumberVar("swordsAmount", 9001)
vm.SetBoolVar("lie",false)
chant := "I am the bone of my sword.."
vm.SetStringVar("chant",chant)

smalltalkProgram1 := `(swordsAmount > 9000) ifTrue:[chant] ifFalse:['ouch it hurts']`
smalltalkProgram2 := `(swordsAmount > 1.2e4) ifTrue:[-1] ifFalse:[42]`
smalltalkProgram3 := `lie ifTrue:[0.001] ifFalse:[-0.56]`
smalltalkProgram4 := `swordsAmount > 1.2e4`

var result1 string
var result2 int64
var result3 float64
var result4 bool

result1 = vm.EvaluateToString(smalltalkProgram1)
testutils.ASSERT_STREQ(t, result1, chant)

result2 = vm.EvaluateToInt64(smalltalkProgram2)
testutils.ASSERT_EQ(t, int(result2), 42)

result3 = vm.EvaluateToFloat64(smalltalkProgram3)
testutils.ASSERT_FLOAT64_EQ(t, result3, -0.56)

result4 = vm.EvaluateToBool(smalltalkProgram4)
testutils.ASSERT_FALSE(t, result4)
```
##### Nested variable scopes
```go
var inputString1, inputString2 string
var result1, result2 int64

vm := NewSmalltalkVM()
vm.SetNumberVar("x", 11)

inputString1 = `|x| x := 25. x+75`
result1 = vm.EvaluateToInt64(inputString1)
testutils.ASSERT_EQ(t, int(result1), 100)

inputString2 = `x+75`
result2 = vm.EvaluateToInt64(inputString2)
testutils.ASSERT_EQ(t, int(result2), 86)
```
