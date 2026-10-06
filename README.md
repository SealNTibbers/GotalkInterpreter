# GotalkInterpreter

It's a simplistic Smalltalk code interpreter written in Golang by Alex and Michael.


#### Who can use it

The entire purpose of this library is to use Smalltalk for dynamic code (string) evaluation in Golang applications. It is optimized to reevaluate same code lines with different scope (variables). Typical use case: our app reads xml file with a markup and a Smalltalk code, evaluates this code and uses the result. We use it to build and animate an OpenGL UI for our embedded software.

#### Why Smalltalk

Smalltalk is beautiful dynamic language with a concise and readable syntax. Also we are smalltalkers so that's why.

#### Contents

Scanner and Parser are essentially standard Smalltalk Scanner and Parser rewritten in Go. We used [VisualWorks](http://www.cincomsmalltalk.com) and [Pharo](https://pharo.org/) realizations as reference implementations.
Evaluator is the API entry point. You can expand functionality by modifying smalltalkObjects.go file.

#### Bugs, Tests etc

There are a lot of tests for everything here so we are pretty sure that this library is actually useable. You can use its tests (specifically smalltalkEvaluator_test.go) to better understand what you can do with this library.
We can rewrite some parts later just to make our Go code better and somehow expand overall functionality.

## Installation

GotalkInterpreter does not use any third party libraries. To get it run on your machine, you just:
```go
go get github.com/SealNTibbers/GotalkInterpreter
```

## Console

`go run .` (or the built binary) is an interactive Smalltalk console. It works like a workspace: every input is
evaluated and its result is printed, and the variables you assign stay defined. Input continues on the next line
(`  >` prompt) while a bracket, string or comment is open or an expression is unfinished. Ctrl+C cancels the
current input, and Ctrl+D or `:quit` exits. `:help` lists the commands (`:load file.st`, `:reset`, `:history`).

```
st> x := 3 + 4.
7
st> Transcript show: 'x squared = '; print: x * x; cr
x squared = 49
```

`gotalk file.st ...` runs files in order and prints only what they write to `Transcript`. A file may be split into
chunks ending with `!` (the classic file-in format, `!!` stands for `!`), may start with a `#!` line, and reports
errors as `file:line: Error: ...`. `gotalk -e 'expr'` prints a result; `gotalk -i file.st` opens the console after
running the file; `gotalk < file.st` reads the program from standard input.

On a terminal, input is edited like in a shell (package `lineedit`, standard library only): Left/Right,
Home/End or Ctrl+A/E, Alt+Left/Right or Alt+B/F by word, Backspace/Delete, Ctrl+K/U/W to delete to the end, to the
start or a word. Up/Down (Ctrl+P/N) recall earlier inputs, and Ctrl+R searches them; an input typed on several lines
is recalled as one line. `:history` lists them. The history is kept in `~/.gotalk_history` (last 1000 inputs);
set `GOTALK_HISTORY` to another file, or to an empty value to keep it in memory. Line editing works on macOS,
Linux and the BSDs; elsewhere the console reads plain lines.

## API and Examples

Result of our Smalltalk code evaluation can be number (float64 or int. Internally it's always float64), bool, string, array or nil.
Characters (`$a`) and symbols (`#foo`) are strings, as in Amber.

In our little Smalltalk we have supported limited amount of messages that is enough for our internal project but it's easily expandable
(see the method tables in `treeNodes/smalltalkObjects.go`).

All objects understand: `=` `~=` `==` `~~` `value` `yourself` `printString` `displayString` `isNil` `notNil`
`ifNil:` `ifNotNil:` `ifNil:ifNotNil:` `ifNotNil:ifNil:` `isNumber` `isString` `isBoolean` `isArray` `isBlock`.

Numbers: `+` `-` `*` `/` `//` `\\` `rem:` `quo:` `<` `<=` `>` `>=` `max:` `min:` `between:and:` `raisedTo:`
`roundTo:` `truncateTo:` `log:` `arcTan:` `abs` `negated` `sqrt` `sqr` `squared` `sin` `cos` `tan` `arcSin` `arcCos` `arcTan`
`ln` `log` `exp` `rounded` `truncated` `floor` `ceiling` `fractionPart` `integerPart` `asInteger` `asFloat` `asNumber`
`sign` `isZero` `even` `odd` `degreesToRadians` `radiansToDegrees` `asString` `printString:` (radix)
`printPaddedWith:to:` `printShowingDecimalPlaces:`.
`\\` and `//` are floored as in Smalltalk (`-7 \\ 2 = 1`, `7.5 \\ 2 = 1.5`). Division by zero is a `ZeroDivide` error.

Booleans: `&` `|` `and:` `or:` `xor:` `not` `ifTrue:` `ifFalse:` `ifTrue:ifFalse:` `ifFalse:ifTrue:` `asString`.

Strings: `,` `<` `<=` `>` `>=` `size` `isEmpty` `notEmpty` `at:` `at:ifAbsent:` `first` `last` `copyFrom:to:`
`includesSubstring:` `reversed` `asUppercase` `asLowercase` `asNumber` `asString` `asSymbol`.

Blocks: `value` `value:` `value:value:` `value:value:value:` `value:value:value:value:` `valueWithArguments:` `numArgs`.

Arrays: `at:` `at:ifAbsent:` `size` `isEmpty` `notEmpty` `first` `last` `includes:` `indexOf:` `,`
and element-wise `+` `-` `*` `/` `\\` `//` with a number or an array of the same size.

#### Errors

Parse and runtime errors (unknown variable, message not understood, wrong argument type, index out of bounds,
division by zero) are returned as Go errors by `Evaluate`, `EvaluateFloat64`, `EvaluateInt64`, `EvaluateString`
and `EvaluateBool`. The older `RunProgram` and `EvaluateTo*` helpers never panic: they return nil or the zero value
on errors.

Results are cached per program until a variable the program reads changes through the evaluator (`SetVar`, ...).
After changing the global scope directly, call `InvalidateCache()`.

#### Concurrency

An evaluator is safe for concurrent use, and evaluations run in parallel: cache hits take no lock, and other
evaluations share a read lock. `SetVar` and friends take the write lock, so they wait for running evaluations and no
program sees a parameter change halfway. `SetVars` changes several parameters as one step. Workspace evaluators
(`NewSmalltalkWorkspace`) run one program at a time, because their programs write shared variables.
Don't modify a value after setting it, and don't change the global scope directly while evaluations run.

`./bench.sh` runs the parallel benchmarks (`evaluator/concurrency_bench_test.go`) for 1–16 goroutines and prints a
table; `./bench.sh old.txt` adds a comparison with an earlier run (`bench-latest.txt` holds the raw output).

#### Low-lewel API example
```go
//so we have smalltalk code string and want to evaluate it `angle\\10/10-0.9*10`

globalScope := new(treeNodes.Scope).Initialize()
globalScope.SetVar("angle", treeNodes.NewSmalltalkNumber(25))

evaluator = evaluator.NewEvaluatorWithGlobalScope(globalScope)
resultObject, err := evaluator.Evaluate(`angle\\10/10-0.9*10`)
if err != nil {
    // parse or runtime error
}

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
