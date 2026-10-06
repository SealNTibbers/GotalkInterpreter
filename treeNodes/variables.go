package treeNodes

import "sort"

// FreeVariables returns the sorted names a program reads that are not bound inside it by temporaries
// (| a |) or block arguments ([:a | ...]). These are the variables its value depends on.
func FreeVariables(node ProgramNodeInterface) []string {
	analysis := analyzeVariables(node)
	return analysis.read
}

// AssignedFreeVariables returns the sorted names a program assigns that are not bound inside it.
// They live outside the program (in a workspace, or as program-local variables).
func AssignedFreeVariables(node ProgramNodeInterface) []string {
	analysis := analyzeVariables(node)
	return analysis.assigned
}

type variableAnalysis struct {
	bound    map[string]int
	reads    map[string]bool
	assigns  map[string]bool
	read     []string
	assigned []string
}

func analyzeVariables(node ProgramNodeInterface) *variableAnalysis {
	analysis := &variableAnalysis{bound: map[string]int{}, reads: map[string]bool{}, assigns: map[string]bool{}}
	analysis.walk(node)
	analysis.read = sortedNames(analysis.reads)
	analysis.assigned = sortedNames(analysis.assigns)
	return analysis
}

func sortedNames(names map[string]bool) []string {
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func (a *variableAnalysis) bind(variables []*VariableNode) {
	for _, variable := range variables {
		a.bound[variable.GetName()]++
	}
}

func (a *variableAnalysis) unbind(variables []*VariableNode) {
	for _, variable := range variables {
		a.bound[variable.GetName()]--
	}
}

func (a *variableAnalysis) walk(node ProgramNodeInterface) {
	switch n := node.(type) {
	case *SequenceNode:
		a.bind(n.temporaries)
		for _, statement := range n.statements {
			a.walk(statement)
		}
		a.unbind(n.temporaries)
	case *BlockNode:
		a.bind(n.arguments)
		if n.body != nil {
			a.walk(n.body)
		}
		a.unbind(n.arguments)
	case *VariableNode:
		if a.bound[n.GetName()] == 0 {
			a.reads[n.GetName()] = true
		}
	case *AssignmentNode:
		if a.bound[n.variable.GetName()] == 0 {
			a.assigns[n.variable.GetName()] = true
		}
		a.walk(n.value)
	case *MessageNode:
		a.walk(n.receiver)
		for _, argument := range n.arguments {
			a.walk(argument)
		}
	case *CascadeNode:
		for _, message := range n.messages {
			a.walk(message)
		}
	}
}
