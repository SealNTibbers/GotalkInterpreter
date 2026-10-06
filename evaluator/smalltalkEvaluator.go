package evaluator

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/SealNTibbers/GotalkInterpreter/parser"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

// testing stuff
func NewTestEvaluator() *Evaluator {
	return NewSmalltalkVM()
}

// TestEval evaluates code without variables. It returns nil when the code fails.
func TestEval(codeString string) treeNodes.SmalltalkObjectInterface {
	result, _ := NewTestEvaluator().Evaluate(codeString)
	return result
}

// TestEvalWithScope evaluates code with the given global variables. It returns nil when the code fails.
func TestEvalWithScope(codeString string, scope *treeNodes.Scope) treeNodes.SmalltalkObjectInterface {
	result, _ := NewEvaluatorWithGlobalScope(scope).Evaluate(codeString)
	return result
}

// real world API

// NewSmalltalkVM returns an evaluator for expressions over global variables. Each program runs in its own
// local scope, and results are cached until a variable they read changes.
func NewSmalltalkVM() *Evaluator {
	globalScope := new(treeNodes.Scope).Initialize()
	return NewEvaluatorWithGlobalScope(globalScope)
}

// NewSmalltalkWorkspace returns an evaluator whose programs share a scope, like a Smalltalk workspace:
// a variable assigned by one program can be read by the next.
func NewSmalltalkWorkspace() *Evaluator {
	evaluator := NewSmalltalkVM()
	evaluator.workspaceScope = new(treeNodes.Scope).Initialize()
	evaluator.workspaceScope.OuterScope = evaluator.globalScope
	return evaluator
}

func NewEvaluatorWithGlobalScope(global *treeNodes.Scope) *Evaluator {
	evaluator := new(Evaluator)
	evaluator.dependents = make(map[string]map[*program]struct{})
	evaluator.globalScope = global.MarkAsGlobal()
	return evaluator
}

// Evaluator parses programs once and caches their results. It is safe for concurrent use: evaluations run in
// parallel, while SetVar and friends wait for running evaluations and block new ones, so a program never sees
// a parameter change halfway through. Workspace evaluators (NewSmalltalkWorkspace) run one program at a time,
// because their programs write the shared workspace scope.
//
// Change the global scope only through the evaluator. If you change it directly (GetGlobalScope), no
// evaluation may run meanwhile, and call InvalidateCache afterwards. Values must not be modified after they
// were set, and Deferred blocks must not assign variables outside themselves: they can run in parallel.
type Evaluator struct {
	// lock guards the scopes and the cache validity. Evaluations hold it for reading, changes for writing.
	lock           sync.RWMutex
	globalScope    *treeNodes.Scope
	workspaceScope *treeNodes.Scope
	// programString -> *program. Parsing happens outside any lock; registering a program takes indexLock.
	programs  sync.Map
	indexLock sync.Mutex
	// variable name -> programs that read it (guarded by indexLock)
	dependents map[string]map[*program]struct{}
}

// program is immutable after compile, except for its cached result.
type program struct {
	node      treeNodes.ProgramNodeInterface
	parseErr  error
	reads     []string
	assigns   []string
	cacheable bool
	// *result; a nil *result means the cache is invalid. Set during evaluations (read lock), cleared under the
	// write lock, so a result computed before a parameter change can't be stored after it.
	cached atomic.Value
}

type result struct {
	value treeNodes.SmalltalkObjectInterface
	err   error
}

var noResult *result

func (e *Evaluator) SetGlobalScope(scope *treeNodes.Scope) *Evaluator {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.globalScope = scope.MarkAsGlobal()
	if e.workspaceScope != nil {
		e.workspaceScope.OuterScope = e.globalScope
	}
	e.invalidateAll()
	return e
}

func (e *Evaluator) GetGlobalScope() *treeNodes.Scope {
	return e.globalScope
}

// InvalidateCache drops all cached results, e.g. after the global scope was changed directly.
func (e *Evaluator) InvalidateCache() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.invalidateAll()
}

// Evaluate runs a program and returns its result, or the parse or runtime error.
func (e *Evaluator) Evaluate(programString string) (treeNodes.SmalltalkObjectInterface, error) {
	p := e.compile(programString)
	// A cache hit needs no lock: results are only invalidated under the write lock, so a valid one is the
	// result for the current parameters, or for the ones before a change that is still being made.
	if cached := p.cached.Load().(*result); cached != nil {
		return cached.value, cached.err
	}
	if e.workspaceScope != nil {
		e.lock.Lock()
		defer e.lock.Unlock()
	} else {
		e.lock.RLock()
		defer e.lock.RUnlock()
	}
	return e.run(p)
}

// EvaluateFloat64 runs a program that must answer a Number.
func (e *Evaluator) EvaluateFloat64(programString string) (float64, error) {
	result, err := e.Evaluate(programString)
	if err != nil {
		return 0, err
	}
	number, ok := result.(*treeNodes.SmalltalkNumber)
	if !ok {
		return 0, fmt.Errorf("expected a Number, got %s", describe(result))
	}
	return number.GetValue(), nil
}

// EvaluateInt64 runs a program that must answer a Number and truncates it.
func (e *Evaluator) EvaluateInt64(programString string) (int64, error) {
	value, err := e.EvaluateFloat64(programString)
	return int64(value), err
}

// EvaluateString runs a program that must answer a String.
func (e *Evaluator) EvaluateString(programString string) (string, error) {
	result, err := e.Evaluate(programString)
	if err != nil {
		return "", err
	}
	str, ok := result.(*treeNodes.SmalltalkString)
	if !ok {
		return "", fmt.Errorf("expected a String, got %s", describe(result))
	}
	return str.GetValue(), nil
}

// EvaluateBool runs a program that must answer a Boolean.
func (e *Evaluator) EvaluateBool(programString string) (bool, error) {
	result, err := e.Evaluate(programString)
	if err != nil {
		return false, err
	}
	boolean, ok := result.(*treeNodes.SmalltalkBoolean)
	if !ok {
		return false, fmt.Errorf("expected a Boolean, got %s", describe(result))
	}
	return boolean.GetValue(), nil
}

func describe(object treeNodes.SmalltalkObjectInterface) string {
	printed, err := object.Perform("printString", nil)
	if err != nil {
		return object.TypeOf()
	}
	return printed.(*treeNodes.SmalltalkString).GetValue()
}

// RunProgram runs a program and returns its result, or nil (UndefinedObject) when it fails.
// Use Evaluate to get the error.
func (e *Evaluator) RunProgram(programString string) treeNodes.SmalltalkObjectInterface {
	result, err := e.Evaluate(programString)
	if err != nil {
		return treeNodes.NewSmalltalkUndefinedObject()
	}
	return result
}

// EvaluateProgram evaluates an already parsed program without caching. It returns nil when it fails.
func (e *Evaluator) EvaluateProgram(node treeNodes.ProgramNodeInterface) treeNodes.SmalltalkObjectInterface {
	if e.workspaceScope != nil {
		e.lock.Lock()
		defer e.lock.Unlock()
	} else {
		e.lock.RLock()
		defer e.lock.RUnlock()
	}
	result, err := node.Eval(e.newLocalScope())
	if err != nil {
		return nil
	}
	return result
}

// The EvaluateTo* helpers return the zero value when the program fails or answers another type.
// Use EvaluateFloat64, EvaluateString, ... to get the error.

func (e *Evaluator) EvaluateToString(programString string) string {
	result, _ := e.EvaluateString(programString)
	return result
}

func (e *Evaluator) EvaluateToFloat64(programString string) float64 {
	result, _ := e.EvaluateFloat64(programString)
	return result
}

func (e *Evaluator) EvaluateToInt64(programString string) int64 {
	result, _ := e.EvaluateInt64(programString)
	return result
}

func (e *Evaluator) EvaluateToBool(programString string) bool {
	result, _ := e.EvaluateBool(programString)
	return result
}

// EvaluateToInterface returns the result as float64, string, bool, []interface{} or nil (also on errors).
func (e *Evaluator) EvaluateToInterface(programString string) interface{} {
	resultObject, err := e.Evaluate(programString)
	if err != nil || resultObject == nil {
		return nil
	}
	switch resultObject.TypeOf() {
	case treeNodes.NUMBER_OBJ:
		return resultObject.(*treeNodes.SmalltalkNumber).GetValue()
	case treeNodes.STRING_OBJ:
		return resultObject.(*treeNodes.SmalltalkString).GetValue()
	case treeNodes.BOOLEAN_OBJ:
		return resultObject.(*treeNodes.SmalltalkBoolean).GetValue()
	case treeNodes.ARRAY_OBJ:
		array, err := resultObject.(*treeNodes.SmalltalkArray).GetValue()
		if err != nil {
			return nil
		}
		return array
	default:
		return nil
	}
}

// newLocalScope holds a run's temporaries. In a workspace, undeclared variables a program assigns go one level
// up, into the shared workspace scope.
func (e *Evaluator) newLocalScope() *treeNodes.Scope {
	scope := new(treeNodes.Scope).Initialize()
	scope.OuterScope = e.globalScope
	if e.workspaceScope != nil {
		scope.OuterScope = e.workspaceScope
	}
	return scope
}

// compile parses a program once; parse errors are kept so a bad formula isn't parsed again on every frame.
// Two goroutines may parse a new program at the same time; the first one registered wins.
func (e *Evaluator) compile(programString string) *program {
	if p, ok := e.programs.Load(programString); ok {
		return p.(*program)
	}
	p := &program{}
	p.cached.Store(noResult)
	p.node, p.parseErr = parser.InitializeParserFor(programString)
	if p.parseErr == nil {
		p.reads = treeNodes.FreeVariables(p.node)
		p.assigns = treeNodes.AssignedFreeVariables(p.node)
		// In a workspace, assignments are side effects other programs see, so the program must run every time.
		p.cacheable = e.workspaceScope == nil || len(p.assigns) == 0
	}
	e.indexLock.Lock()
	defer e.indexLock.Unlock()
	if existing, loaded := e.programs.LoadOrStore(programString, p); loaded {
		return existing.(*program)
	}
	for _, name := range p.reads {
		if e.dependents[name] == nil {
			e.dependents[name] = make(map[*program]struct{})
		}
		e.dependents[name][p] = struct{}{}
	}
	return p
}

// run evaluates a program, holding e.lock for reading (for writing in a workspace).
func (e *Evaluator) run(p *program) (treeNodes.SmalltalkObjectInterface, error) {
	if p.parseErr != nil {
		return nil, p.parseErr
	}
	if cached := p.cached.Load().(*result); cached != nil {
		return cached.value, cached.err
	}
	value, err := p.node.Eval(e.newLocalScope())
	if p.cacheable && !e.readsDeferred(p) {
		p.cached.Store(&result{value, err})
	}
	if e.workspaceScope != nil {
		for _, name := range p.assigns {
			e.invalidate(name)
		}
	}
	return value, err
}

// readsDeferred tells whether a program reads a Deferred global, whose value can change at any time.
func (e *Evaluator) readsDeferred(p *program) bool {
	for _, name := range p.reads {
		if value, err := e.globalScope.GetVarValue(name); err == nil && value != nil && value.TypeOf() == treeNodes.DEFERRED {
			return true
		}
	}
	return false
}

// invalidate and invalidateAll need e.lock held for writing.
func (e *Evaluator) invalidate(variableName string) {
	e.indexLock.Lock()
	defer e.indexLock.Unlock()
	for p := range e.dependents[variableName] {
		p.cached.Store(noResult)
	}
}

func (e *Evaluator) invalidateAll() {
	e.programs.Range(func(_, p interface{}) bool {
		p.(*program).cached.Store(noResult)
		return true
	})
}

// scope-related delegations
func (e *Evaluator) SetVar(name string, value treeNodes.SmalltalkObjectInterface) treeNodes.SmalltalkObjectInterface {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.invalidate(name)
	return e.globalScope.SetVar(name, value)
}

func (e *Evaluator) SetStringVar(name string, value string) treeNodes.SmalltalkObjectInterface {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.invalidate(name)
	return e.globalScope.SetStringVar(name, value)
}

func (e *Evaluator) SetNumberVar(name string, value float64) treeNodes.SmalltalkObjectInterface {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.invalidate(name)
	return e.globalScope.SetNumberVar(name, value)
}

func (e *Evaluator) SetBoolVar(name string, value bool) treeNodes.SmalltalkObjectInterface {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.invalidate(name)
	return e.globalScope.SetBoolVar(name, value)
}

// SetVars sets several variables at once: no evaluation sees some of them changed and others not.
func (e *Evaluator) SetVars(values map[string]treeNodes.SmalltalkObjectInterface) {
	e.lock.Lock()
	defer e.lock.Unlock()
	for name, value := range values {
		e.invalidate(name)
		e.globalScope.SetVar(name, value)
	}
}

func (e *Evaluator) FindValueByName(name string) (treeNodes.SmalltalkObjectInterface, bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.globalScope.FindValueByName(name)
}
