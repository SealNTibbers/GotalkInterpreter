package evaluator

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

// printed evaluates code and returns its printString, or "error: ..." when it fails.
func printed(vm *Evaluator, code string) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = "panic"
		}
	}()
	value, err := vm.Evaluate(code)
	if err != nil {
		return "error: " + err.Error()
	}
	printString, err := value.Perform("printString", nil)
	if err != nil {
		return "error: " + err.Error()
	}
	return printString.(*treeNodes.SmalltalkString).GetValue()
}

func checkAll(t *testing.T, vm *Evaluator, cases [][2]string) {
	t.Helper()
	for _, c := range cases {
		got := printed(vm, c[0])
		if strings.HasPrefix(c[1], "error: ") {
			if !strings.HasPrefix(got, "error: ") || !strings.Contains(got, strings.TrimPrefix(c[1], "error: ")) {
				t.Errorf("%s: expected %q, got %q", c[0], c[1], got)
			}
		} else if got != c[1] {
			t.Errorf("%s: expected %q, got %q", c[0], c[1], got)
		}
	}
}

func TestBlockWithUnusedArgumentDoesNotBreakCache(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("speed", 120)
	checkAll(t, vm, [][2]string{{`[:x | speed] value: 1`, `120`}})
	vm.SetNumberVar("speed", 10) // used to panic in BlockNode.GetVariables
	checkAll(t, vm, [][2]string{{`[:x | speed] value: 1`, `10`}})
}

func TestNoPanics(t *testing.T) {
	vm := NewSmalltalkVM()
	checkAll(t, vm, [][2]string{
		{`#(1 2 3) at: 7`, `error: out of bounds`},
		{`#(1 2 3) at: 0`, `error: out of bounds`},
		{`#(1 2 3) at: 1.5`, `error: Integer`},
		{`#(1 $a sym #sym foo: #+ + at:put: true nil 'str' (2 3) #(4))`, `#(1 'a' 'sym' 'sym' 'foo:' '+' '+' 'at:put:' true nil 'str' #(2 3) #(4))`},
		{`3 + 'a'`, `error: expects a Number`},
		{`true + 1`, `error: Boolean does not understand #+`},
		{`[:a | a] value`, `error: wrong number of arguments`},
		{`[1] value: 2`, `error: wrong number of arguments`},
		{`true and: [3]`, `3`},
		{`#('a') * 2`, `error: not a Number`},
		{`1 +`, `error: expression expected`},
		{`foo:`, `error: expression expected`},
		{`3 -`, `error: expression expected`},
		{`'abc`, `error: UnmatchedQuote`},
		{`3 "unterminated`, `error: unterminated comment`},
		{`3 ]`, `error: unexpected ']'`},
		{`(3))`, `error: unexpected ')'`},
		{`{1. 2}`, `error: unexpected character`},
		{`#`, `error: literal expected`},
		{`$`, `error: character expected`},
	})
}

func TestErrorsAreReported(t *testing.T) {
	vm := NewSmalltalkVM()
	if _, err := vm.Evaluate(`x + 1`); err == nil || !strings.Contains(err.Error(), "undefined variable: x") {
		t.Errorf("unknown variable: %v", err)
	}
	if _, err := vm.Evaluate(`1 / 0`); !errors.Is(err, treeNodes.ErrZeroDivide) {
		t.Errorf("1 / 0: %v", err)
	}
	if _, err := vm.EvaluateFloat64(`'text'`); err == nil {
		t.Error("EvaluateFloat64 on a String should fail")
	}
	// the convenience helpers no longer panic
	if vm.EvaluateToFloat64(`'text'`) != 0 || vm.EvaluateToString(`1 +`) != "" || vm.EvaluateToBool(`x`) {
		t.Error("EvaluateTo* should return zero values on errors")
	}
	if vm.RunProgram(`x`).TypeOf() != treeNodes.UNDEFINED_OBJ {
		t.Error("RunProgram should return nil on errors")
	}
	// a variable defined after a failed run is picked up
	vm.SetNumberVar("x", 2)
	checkAll(t, vm, [][2]string{{`x + 1`, `3`}})
}

func TestSmalltalkArithmetic(t *testing.T) {
	vm := NewSmalltalkVM()
	checkAll(t, vm, [][2]string{
		{`-7 \\ 2`, `1`},
		{`7 \\ -2`, `-1`},
		{`7.5 \\ 2`, `1.5`},
		{`370 \\ 360.5`, `9.5`},
		{`-10 \\ 360`, `350`},
		{`-7 // 2`, `-4`},
		{`-7 rem: 2`, `-1`},
		{`-7 quo: 2`, `-3`},
		{`5 // 0`, `error: ZeroDivide`},
		{`5 \\ 0`, `error: ZeroDivide`},
		{`2 raisedTo: 10`, `1024`},
		{`5 between: 1 and: 10`, `true`},
		{`12.345 roundTo: 0.5`, `12.5`},
		{`17 truncateTo: 5`, `15`},
		{`3 - -2`, `5`},
		{`3-2.345`, `0.6549999999999998`}, // used to be rounded to 3 - 2.35,
		{`10-0.125`, `9.875`},
		{`16r1F`, `31`},
		{`2r1010`, `10`},
		{`1.5e2`, `150`},
		{`0.1 + 0.2`, `0.30000000000000004`},
		{`-2.5 rounded`, `-3`},
		{`-3 sign`, `-1`},
		{`4 even`, `true`},
		{`1 arcTan: 1`, `0.7853981633974483`},
		{`3 "a comment" + 4`, `7`},
	})
}

func TestStringsAndPrinting(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("speed", 120.456)
	checkAll(t, vm, [][2]string{
		{`'SPD ', (speed printShowingDecimalPlaces: 1)`, `'SPD 120.5'`},
		{`7 printPaddedWith: $0 to: 3`, `'007'`},
		{`-5 printPaddedWith: $0 to: 3`, `'-05'`},
		{`255 printString: 16`, `'FF'`},
		{`3 printString`, `'3'`},
		{`2.5 printString`, `'2.5'`},
		{`1e-8 printString`, `'1e-8'`},
		{`'it''s' printString`, `'''it''''s'''`},
		{`'abc' size`, `3`},
		{`'abc' at: 2`, `'b'`},
		{`'abc' at: 5 ifAbsent: ['-']`, `'-'`},
		{`'hello' copyFrom: 2 to: 4`, `'ell'`},
		{`'abc' reversed asUppercase`, `'CBA'`},
		{`'12.5' asNumber + 1`, `13.5`},
		{`'a' < 'b'`, `true`},
		{`'a' = 'a'`, `true`},
		{`'a' = 1`, `false`},
		{`'a', 1`, `error: expects a String`},
		{`$a`, `'a'`},
		{`#foo:bar:`, `'foo:bar:'`},
	})
}

func TestNilAndObjects(t *testing.T) {
	vm := NewSmalltalkVM()
	checkAll(t, vm, [][2]string{
		{`nil`, `nil`},
		{`nil isNil`, `true`},
		{`3 isNil`, `false`},
		{`nil ifNil: [1] ifNotNil: [:x | x]`, `1`},
		{`5 ifNil: [1] ifNotNil: [:x | x * 2]`, `10`},
		{`false ifTrue: [1]`, `nil`},
		{`[] value`, `nil`},
		{`| a | a`, `nil`},
		{`#(1 2) = #(1 2)`, `true`},
		{`3 == 3`, `true`},
		{`3 printString; + 4`, `7`},
	})
}

func TestBlocksAndArrays(t *testing.T) {
	vm := NewSmalltalkVM()
	checkAll(t, vm, [][2]string{
		{`[:a :b | a + b] value: 1 value: 2`, `3`},
		{`[:a :b :c | a + b + c] value: 1 value: 2 value: 3`, `6`},
		{`[:a :b | a * b] valueWithArguments: #(3 4)`, `12`},
		{`[:a :b | a] numArgs`, `2`},
		// the editor appends `at: i ifAbsent: [0]` for array formulas in repeating groups
		{`#(0 10 20) at: 2 ifAbsent: [0]`, `10`},
		{`#(0 10 20) at: 4 ifAbsent: [0]`, `0`},
		{`[:index | #('0' '10' '20') at: index] value: 3`, `'20'`},
		{`#(1 2 3) size`, `3`},
		{`#(1 2 3) includes: 2`, `true`},
		{`#(1 2) , #(3)`, `#(1 2 3)`},
		{`#(1 2) + #(10 20)`, `#(11 22)`},
		{`#(-1 -2) * 2`, `#(-2 -4)`},
	})
}

func TestScopes(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("x", 11)
	checkAll(t, vm, [][2]string{
		// assignment inside a block writes the enclosing temporary (used to create a block-local copy)
		{`| a | a := 1. [:v | a := v] value: 5. a`, `5`},
		{`| a | a := 1. [a := 2] value. a`, `2`},
		// block temporaries stay inside the block
		{`| a | a := 1. [| a | a := 2] value. a`, `1`},
		// programs shadow parameters instead of overwriting them
		{`x := 3. x`, `3`},
		{`x`, `11`},
	})
}

func TestCacheInvalidation(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("speed", 1)
	checkAll(t, vm, [][2]string{{`| a | a := speed. a * 2`, `2`}})
	vm.SetNumberVar("speed", 5) // an assignment used to hide the dependency on speed
	checkAll(t, vm, [][2]string{{`| a | a := speed. a * 2`, `10`}})

	workspace := NewSmalltalkWorkspace()
	checkAll(t, workspace, [][2]string{
		{`y := 5`, `5`},
		{`y + 1`, `6`},
		{`y := 7`, `7`},
		{`y + 1`, `8`}, // used to return the cached 6
		{`| t | t := 1. t`, `1`},
		{`t`, `error: undefined variable: t`}, // temporaries don't leak into the workspace
	})
}

func TestConcurrentUse(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("speed", 1)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				vm.SetNumberVar("speed", float64(j))
				if _, err := vm.EvaluateFloat64(`speed * 2 + 1`); err != nil {
					t.Error(err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestCascades(t *testing.T) {
	checkAll(t, NewSmalltalkVM(), [][2]string{
		{`3 + 4; * 10`, `30`},
		// unary and keyword parts used to panic or fail to parse
		{`3 printString; yourself`, `3`},
		{`#(1 2) at: 1; at: 2`, `2`},
		{`3 max: 4; min: 1`, `1`},
		{`3 + 4; printString`, `'3'`}, // every part goes to the receiver of the first message
		{`#(5 6) first printString; size`, `error: does not understand #size`},
		{`3 foo; ; bar`, `error: message expected`},
	})
}
