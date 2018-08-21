package treeNodes

import (
	"github.com/SealNTibbers/GotalkInterpreter/scanner"
)

type ProgramNodeInterface interface {
	TypeOfNode() string
	SetParent(nodeInterface ProgramNodeInterface)
	GetParent() ProgramNodeInterface

	IsMessage() bool
	IsLiteralNode() bool
	IsLiteralArray() bool
	IsAssignment() bool
	Eval(scope *Scope) SmalltalkObjectInterface
}

type Node struct {
	parent ProgramNodeInterface
}

func (n *Node) Eval(scope *Scope) SmalltalkObjectInterface {
	return nil
}

func (n *Node) IsMessage() bool {
	return false
}

func (n *Node) IsAssignment() bool {
	return false
}

func (n *Node) IsLiteralNode() bool {
	return false
}

func (l *Node) IsLiteralArray() bool {
	return false
}

func (n *Node) SetParent(nodeInterface ProgramNodeInterface) {
	n.parent = nodeInterface
}

func (n *Node) GetParent() ProgramNodeInterface {
	return n.parent
}

func (n *Node) TypeOfNode() string {
	return "Node"
}

type SequenceNode struct {
	*Node

	leftBar     int64
	rightBar    int64
	temporaries []*VariableNode
	periods     []int64
	statements  []ProgramNodeInterface
}

func (n *SequenceNode) TypeOfNode() string {
	return "SequenceNode"
}

func (m *SequenceNode) SetTemporaries(temporaries []*VariableNode) {
	m.temporaries = temporaries
	for _, temp := range temporaries {
		temp.SetParent(m)
	}
}

func (m *SequenceNode) SetStatements(statements []ProgramNodeInterface) {
	m.statements = statements
	for _, statement := range statements {
		statement.SetParent(m)
	}
}

func (m *SequenceNode) GetStatements() []ProgramNodeInterface {
	return m.statements
}

func (m *SequenceNode) GetTemporaries() []*VariableNode {
	return m.temporaries
}

func (m *SequenceNode) SetPeriods(periods []int64) {
	m.periods = periods
}

func (m *SequenceNode) SetLeftBar(leftBar int64) {
	m.leftBar = leftBar
}

func (m *SequenceNode) SetRightBar(rightBar int64) {
	m.rightBar = rightBar
}

type ValueNodeInterface interface {
	ProgramNodeInterface
	AddParenthesis(interval Interval)
}

type ValueNode struct {
	*Node
	parentheses []Interval
}

func (v *ValueNode) AddParenthesis(interval Interval) {
	v.parentheses = append(v.parentheses, interval)
}

type AssignmentNode struct {
	*ValueNode

	position int64
	variable *VariableNode
	value    ValueNodeInterface
}

func (a *AssignmentNode) GetVariable() *VariableNode {
	return a.variable
}

func (a *AssignmentNode) GetValue() ValueNodeInterface {
	return a.value
}

func (a *AssignmentNode) SetVariable(variable *VariableNode) {
	a.variable = variable
	a.variable.SetParent(a)
}

func (a *AssignmentNode) SetValue(value ValueNodeInterface) {
	a.value = value
	a.value.SetParent(a)
}

func (a *AssignmentNode) SetPosition(position int64) {
	a.position = position
}

func (a *AssignmentNode) IsAssignment() bool {
	return true
}

type LiteralNodeInterface interface {
	ValueNodeInterface
	LiteralToken(token scanner.LiteralTokenInterface) LiteralNodeInterface
	GetValue() string
}

type LiteralNode struct {
	*ValueNode
}

func (l *LiteralNode) IsLiteralNode() bool {
	return true
}

func (l *LiteralNode) LiteralToken(token scanner.LiteralTokenInterface) LiteralNodeInterface {
	if token.TypeOfToken() == scanner.ARRAY {
		return createLiteralArrayNodeFromToken(token)
	} else {
		return &LiteralValueNode{NewLiteralNode(), token}
	}
}

func createLiteralArrayNodeFromToken(token scanner.LiteralTokenInterface) LiteralNodeInterface {
	startPosition := token.GetStart()
	stopPosition := token.GetStop()
	var contents []LiteralNodeInterface
	//TODO: we should fill contents for LiteralArrayNode
	return &LiteralArrayNode{&LiteralNode{NewValueNode()}, startPosition, stopPosition, contents}
}

func CreateLiteralArrayNode(startPosition int64, stopPosition int64, contents []LiteralNodeInterface) LiteralNodeInterface {
	//TODO: we should fill contents for LiteralArrayNode
	node := new(LiteralArrayNode)
	node.LiteralNode = NewLiteralNode()
	node.start = startPosition
	node.stop = stopPosition
	node.contents = contents
	for _, cont := range node.contents {
		cont.SetParent(node)
	}
	return node
}

type LiteralArrayNode struct {
	*LiteralNode
	start    int64
	stop     int64
	contents []LiteralNodeInterface
}

func (l *LiteralArrayNode) IsLiteralArray() bool {
	return true
}

func (l *LiteralArrayNode) GetValue() string {
	var value string
	var separator string
	for i, each := range l.contents {
		if i == 0 {
			separator = ""
		} else {
			separator = " "
		}
		value = value + separator + each.GetValue()
	}
	return value
}

type LiteralValueNode struct {
	*LiteralNode
	token scanner.LiteralTokenInterface
}

func (literalValue *LiteralValueNode) GetTypeOfToken() string {
	return literalValue.token.TypeOfToken()
}

func (literalValue *LiteralValueNode) GetValue() string {
	return literalValue.token.ValueOfToken()
}

type VariableNode struct {
	*ValueNode
	Token scanner.ValueTokenInterface
}

func (v *VariableNode) GetName() string {
	return v.Token.ValueOfToken()
}

type NodeWithRreceiverInterface interface {
	GetReceiver() ValueNodeInterface
}

type MessageNode struct {
	*ValueNode
	receiver      ValueNodeInterface
	selector      *scanner.KeywordToken
	selectorParts []scanner.ValueTokenInterface
	arguments     []ValueNodeInterface
}

func (m *MessageNode) GetReceiver() ValueNodeInterface {
	return m.receiver
}

func (m *MessageNode) GetSelector() string {
	selector := ""
	for _, each := range m.selectorParts {
		selector = selector + each.ValueOfToken()
	}
	return selector
}

func (m *MessageNode) GetSelectorParts() []scanner.ValueTokenInterface {
	return m.selectorParts
}

func (m *MessageNode) GetArguments() []ValueNodeInterface {
	return m.arguments
}

func (n *MessageNode) IsMessage() bool {
	return true
}

func (m *MessageNode) SetReceiverSelectorPartsArguments(receiver ValueNodeInterface, selectorParts []scanner.ValueTokenInterface, arguments []ValueNodeInterface) {
	m.SetReceiver(receiver)
	m.selectorParts = selectorParts
	m.SetArguments(arguments)
}

func (m *MessageNode) SetReceiver(receiver ValueNodeInterface) {
	m.receiver = receiver
	m.receiver.SetParent(m)
}

func (m *MessageNode) SetArguments(arguments []ValueNodeInterface) {
	m.arguments = arguments
	for _, arg := range arguments {
		arg.SetParent(m)
	}
}

type CascadeNode struct {
	*ValueNode
	semicolons []int64
	messages   []*MessageNode
}

func (m *CascadeNode) SetSemicolons(semicolons []int64) {
	m.semicolons = semicolons
}

func (c *CascadeNode) GetReceiver() ValueNodeInterface {
	return c.messages[0].GetReceiver()
}

func (m *CascadeNode) SetMessages(messages []*MessageNode) {
	m.messages = messages
	for _, message := range messages {
		message.SetParent(m)
	}
}

type BlockNode struct {
	*ValueNode
	bar       int64
	arguments []*VariableNode
	colons    []int64
	left      int64
	right     int64
	body      *SequenceNode
}

func (m *BlockNode) GetBody() *SequenceNode {
	return m.body
}

func (m *BlockNode) SetArguments(arguments []*VariableNode) {
	m.arguments = arguments
	for _, arg := range arguments {
		arg.SetParent(m)
	}
}

func (m *BlockNode) SetColons(colons []int64) {
	m.colons = colons
}

func (m *BlockNode) SetBody(body *SequenceNode) {
	m.body = body
}

func (m *BlockNode) SetBar(bar int64) {
	m.bar = bar
}

func (m *BlockNode) SetLeft(left int64) {
	m.left = left
}

func (m *BlockNode) SetRight(right int64) {
	m.right = right
}

type Interval struct {
	start int64
	stop  int64
}

func (i *Interval) SetStart(start int64) {
	i.start = start
}

func (i *Interval) SetStop(stop int64) {
	i.stop = stop
}

func NewAssignmentNode() *AssignmentNode {
	node := new(AssignmentNode)
	node.ValueNode = NewValueNode()
	return node
}

func NewLiteralNode() *LiteralNode {
	node := new(LiteralNode)
	node.ValueNode = NewValueNode()
	return node
}

func NewBlockNode() *BlockNode {
	node := new(BlockNode)
	node.ValueNode = NewValueNode()
	return node
}

func NewMessageNode() *MessageNode {
	node := new(MessageNode)
	node.ValueNode = NewValueNode()
	return node
}

func NewCascadeNode() *CascadeNode {
	node := new(CascadeNode)
	node.ValueNode = NewValueNode()
	return node
}

func NewValueNode() *ValueNode {
	node := new(ValueNode)
	node.Node = &Node{}
	return node
}

func NewSequenceNode() *SequenceNode {
	node := new(SequenceNode)
	node.Node = &Node{}
	return node
}
