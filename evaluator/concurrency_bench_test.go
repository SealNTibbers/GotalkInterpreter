package evaluator

import (
	"sync/atomic"
	"testing"

	"github.com/SealNTibbers/GotalkInterpreter/parser"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

// Benchmarks for evaluating from many goroutines. Run them with several GOMAXPROCS values and compare:
//
//	./bench.sh                      # table for the current code
//	./bench.sh old.txt              # the same, with a column comparing against a saved run
//
// Each benchmark evaluates gaugeFormulas round robin; ns/op is the time per formula.

// gaugeFormulas look like the property scripts of an MFD gauge.
var gaugeFormulas = []string{
	`angle * 2 + 10`,
	`(angle degreesToRadians sin * radius) rounded`,
	`(angle degreesToRadians cos * radius) + centerX`,
	`speed > 250 ifTrue: [1] ifFalse: [2]`,
	`(speed - 40 max: 0) / 3.6`,
	`#('0' '10' '20' '30') at: (speed \\ 4) + 1`,
	`(altitude // 100) printString , ' ft'`,
	`altitude printPaddedWith: $0 to: 6`,
	`| a b | a := speed * 2. b := a + altitude. b / 10`,
	`speed isNil ifTrue: [0] ifFalse: [speed negated abs]`,
	`[:x :y | x * y + angle] value: speed value: 3`,
	`#(1 2 3 4) * angle`,
	`(speed between: 100 and: 300) & (altitude > 1000)`,
	`((angle \\ 360) / 360 * 100) truncated`,
	`radius sqrt + (centerX raisedTo: 2) ln`,
	`altitude > 10000 ifTrue: ['FL' , (altitude // 100) printString] ifFalse: [altitude printString]`,
}

func newGaugeVM(b *testing.B) *Evaluator {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("angle", 37)
	vm.SetNumberVar("radius", 120)
	vm.SetNumberVar("centerX", 400)
	vm.SetNumberVar("speed", 180)
	vm.SetNumberVar("altitude", 12500)
	for _, formula := range gaugeFormulas {
		if _, err := vm.Evaluate(formula); err != nil {
			b.Fatalf("%s: %v", formula, err)
		}
	}
	return vm
}

// deferredNumber returns a global whose value is computed on every read, so programs reading it are never cached.
func deferredNumber(b *testing.B, code string) *treeNodes.Deferred {
	node, err := parser.InitializeParserFor(code)
	if err != nil {
		b.Fatal(err)
	}
	block, ok := node.(*treeNodes.BlockNode)
	if !ok {
		b.Fatalf("%s is not a block", code)
	}
	return treeNodes.NewDeferred(block, new(treeNodes.Scope).Initialize())
}

func evaluateRoundRobin(b *testing.B, vm *Evaluator, pb *testing.PB, start uint32, writeEvery int) {
	i := int(start)
	for pb.Next() {
		if writeEvery > 0 && i%writeEvery == 0 {
			vm.SetNumberVar("speed", float64(i%400))
		}
		if _, err := vm.Evaluate(gaugeFormulas[i%len(gaugeFormulas)]); err != nil {
			b.Error(err)
			return
		}
		i++
	}
}

// Parameters don't change: every evaluation is a cache hit.
func BenchmarkParallelCached(b *testing.B) {
	vm := newGaugeVM(b)
	var seed uint32
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		evaluateRoundRobin(b, vm, pb, atomic.AddUint32(&seed, 7), 0)
	})
}

// The parameters are Deferred, so every evaluation runs the program.
func BenchmarkParallelUncached(b *testing.B) {
	vm := newGaugeVM(b)
	vm.SetVar("angle", deferredNumber(b, `[37]`))
	vm.SetVar("speed", deferredNumber(b, `[180]`))
	var seed uint32
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		evaluateRoundRobin(b, vm, pb, atomic.AddUint32(&seed, 7), 0)
	})
}

// One parameter change every 16 evaluations, like a running display; a change invalidates the formulas reading it.
func BenchmarkParallelMixed(b *testing.B) {
	vm := newGaugeVM(b)
	var seed uint32
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		evaluateRoundRobin(b, vm, pb, atomic.AddUint32(&seed, 7), 16)
	})
}

// One evaluator per goroutine, nothing shared: the upper bound for Uncached.
func BenchmarkParallelIsolated(b *testing.B) {
	var seed uint32
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		vm := newGaugeVM(b)
		vm.SetVar("angle", deferredNumber(b, `[37]`))
		vm.SetVar("speed", deferredNumber(b, `[180]`))
		evaluateRoundRobin(b, vm, pb, atomic.AddUint32(&seed, 7), 0)
	})
}
