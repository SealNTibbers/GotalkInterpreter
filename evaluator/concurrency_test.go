package evaluator

import (
	"fmt"
	"sync"
	"testing"

	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

// Run with -race (./test.sh does).

// Readers never see a half-applied SetVars, and never get a result cached from before a change.
func TestConcurrentSnapshotAndCache(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetVars(map[string]treeNodes.SmalltalkObjectInterface{
		"a": treeNodes.NewSmalltalkNumber(0), "b": treeNodes.NewSmalltalkNumber(0)})
	const writes = 500
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if diff, err := vm.EvaluateFloat64(`a - b`); err != nil || diff != 0 {
					t.Errorf("a - b = %v, %v", diff, err)
					return
				}
				vm.Evaluate(`a * 2`)
			}
		}()
	}
	for i := 1; i <= writes; i++ {
		vm.SetVars(map[string]treeNodes.SmalltalkObjectInterface{
			"a": treeNodes.NewSmalltalkNumber(float64(i)), "b": treeNodes.NewSmalltalkNumber(float64(i))})
		if value, _ := vm.EvaluateFloat64(`a * 2`); value != float64(2*i) {
			t.Fatalf("after setting a to %d, a * 2 = %v", i, value)
		}
	}
	close(stop)
	wg.Wait()
}

// Many goroutines compiling the same new programs at once all get correct, invalidatable results.
func TestConcurrentCompile(t *testing.T) {
	vm := NewSmalltalkVM()
	vm.SetNumberVar("x", 1)
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				program := fmt.Sprintf(`x + %d`, i)
				if value, err := vm.EvaluateFloat64(program); err != nil || value != float64(1+i) {
					t.Errorf("%s = %v, %v", program, value, err)
				}
			}
		}()
	}
	wg.Wait()
	vm.SetNumberVar("x", 100)
	for i := 0; i < 50; i++ {
		if value, _ := vm.EvaluateFloat64(fmt.Sprintf(`x + %d`, i)); value != float64(100+i) {
			t.Fatalf("x + %d = %v after x changed", i, value)
		}
	}
}

// Workspace programs share variables, so they run one at a time.
func TestConcurrentWorkspace(t *testing.T) {
	workspace := NewSmalltalkWorkspace()
	workspace.Evaluate(`counter := 0`)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				workspace.Evaluate(`counter := counter + 1`)
			}
		}()
	}
	wg.Wait()
	if value, _ := workspace.EvaluateFloat64(`counter`); value != 800 {
		t.Fatalf("counter = %v, want 800", value)
	}
}
