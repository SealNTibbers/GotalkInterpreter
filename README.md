# GotalkInterpreter

It's a simplistic Smalltalk code interpreter library written on Golang by Alex and Michael.


#### Who can use it

The entire purpose of this library is to use smalltalk as scripting language in any Golang application since Golang is static by its nature. 

#### Why Smalltalk

Smalltalk is beautiful dynamic language with readable syntax. Also we are smalltalkers with up to 10 years coding experience.

#### Contents

Scanner and Parser are essentially standard Smalltalk Scanner and Parser rewritten in Go. We used [VisualWorks](http://www.cincomsmalltalk.com) and [Pharo](https://pharo.org/) realizations as reference implementations.
Evaluator is the API entry point. You can expand functionality by modifying smalltalkObjects.go file.

#### Bugs, Tests etc

There are a lot of tests for everything here so we are pretty sure that this library is actually useable. You can use this tests (specifically smalltalkEvaluator_test.go) to better understand what you can do with this library.
We can rewrite some parts later just to make our Go code better and somehow expand overall functionality.

## Installation

GotalkInterpreter does not use any third party libraries. For getting it to run on your machine, you just run standard go get:
```go
go get github.com/SealNTibbers/GotalkInterpreter
```

## API and Examples

result of our Smalltalk code evaluation can be number (float64 or int. Internally it's always float64), bool or string

#### Low-lewel API example
```
//so we have smalltalk code string and want to evaluate it `angle\\10/10-0.9*10`

globalScope := new(treeNodes.Scope).Initialize()
globalScope.SetVar("angle", treeNodes.NewSmalltalkNumber(25))

evaluator = evaluator.NewEvaluatorWithGlobalScope(globalScope)
resultObject = evaluator.RunProgram(`angle\\10/10-0.9*10`)

result = resultObject.(*treeNodes.SmalltalkNumber).GetValue()
//so now we have result which is float64 and equals to -4
```

#### Higher level API example. See TestAPI func in smalltalkEvaluator_test.go
```
chant := "I am the bone of my sword.."
globalScope := new(treeNodes.Scope).Initialize()
globalScope.SetNumberVar("swordsAmount", 9001)
globalScope.SetBoolVar("lie",false)
globalScope.SetStringVar("chant",chant)
evaluator := NewEvaluatorWithGlobalScope(globalScope)
smalltalkProgramm1 := `(swordsAmount > 9000) ifTrue:[chant] ifFalse:['ouch it hurts']`
smalltalkProgramm2 := `(swordsAmount > 1.2e4) ifTrue:[-1] ifFalse:[42]`
smalltalkProgramm3 := `lie ifTrue:[0.001] ifFalse:[-0.56]`
smalltalkProgramm4 := `swordsAmount > 1.2e4`
var result1 string
var result2 int64
var result3 float64
var result4 bool

result1 = evaluator.EvaluateToString(smalltalkProgramm1)
testutils.ASSERT_STREQ(t, result1, chant)

result2 = evaluator.EvaluateToInt64(smalltalkProgramm2)
testutils.ASSERT_EQ(t, int(result2), 42)

result3 = evaluator.EvaluateToFloat64(smalltalkProgramm3)
testutils.ASSERT_FLOAT64_EQ(t, result3, -0.56)

result4 = evaluator.EvaluateToBool(smalltalkProgramm4)
testutils.ASSERT_FALSE(t, result4)
```
