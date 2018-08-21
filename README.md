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

## API and Examples

result of our Smalltalk code evaluation can be either float64 or string

```
//so we have smalltalk code string and want to evaluate it `angle\\10/10-0.9*10`

globalScope := new(treeNodes.Scope).Initialize()
globalScope.SetVar("angle", treeNodes.NewSmalltalkNumber(25))

evaluator = evaluator.NewEvaluatorWithGlobalScope(globalScope)
resultObject = evaluator.RunProgram(`angle\\10/10-0.9*10`)

result = resultObject.(*treeNodes.SmalltalkNumber).GetValue()
//so now we have result which is float64 and equals to -4
```
