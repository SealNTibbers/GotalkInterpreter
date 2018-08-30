package parser

import (
	"github.com/SealNTibbers/GotalkInterpreter/io"
	"github.com/SealNTibbers/GotalkInterpreter/scanner"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
	"strconv"
	"strings"
)

type Parser struct {
	scanner         *scanner.Scanner
	currentToken    scanner.TokenInterface
	peekToken       scanner.TokenInterface
	emptyStatements bool
}

func InitializeParserFor(expressionString string) treeNodes.ProgramNodeInterface {
	reader := io.NewReader(expressionString)
	scanner := scanner.New(*reader)
	parser := &Parser{scanner, nil, nil, false}

	//initialize struct members
	parser.step()
	node := parser.parseExpression()
	if len(node.GetStatements()) == 1 && len(node.GetTemporaries()) == 0 {
		return node.GetStatements()[0]
	} else {
		return node
	}
}

func (p *Parser) parseExpression() *treeNodes.SequenceNode {
	node := p.parseStatements(false)
	return node
}

func (p *Parser) parseStatements(tagBool bool) *treeNodes.SequenceNode {
	var leftBar int64
	var rightBar int64
	var args []*treeNodes.VariableNode
	if p.currentToken.IsBinary() {
		if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "|" {
			leftBar = p.currentToken.GetStart()
			p.step()
			args = p.parseArgs()
			if !(p.currentToken.IsBinary() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "|") {
				panic("Parse error in parseStatements function.")
			}
			rightBar = p.currentToken.GetStart()
			p.step()
		} else {
			if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "||" {
				leftBar = p.currentToken.GetStart()
				rightBar = leftBar + 1
				p.step()
			}
		}
	}
	node := treeNodes.NewSequenceNode()
	node.SetLeftBar(leftBar)
	node.SetRightBar(rightBar)
	node.SetTemporaries(args)
	return p.parseStatementListInto(tagBool, node)
}

func (p *Parser) parseArgs() []*treeNodes.VariableNode {
	var args []*treeNodes.VariableNode
	for p.currentToken.IsIdentifier() {
		args = append(args, p.parseVariableNode())
	}
	return args
}

func (p *Parser) parseVariableNode() *treeNodes.VariableNode {
	if p.currentToken.IsIdentifier() {
		return p.parsePrimitiveIdentifier()
	} else {
		panic("we expect variable name here btw")
	}
}

func (p *Parser) parseStatementListInto(tagBool bool, sequenceNode *treeNodes.SequenceNode) *treeNodes.SequenceNode {
	returnFlag := false
	var periods []int64
	var statements []treeNodes.ProgramNodeInterface
	if tagBool {
		//TODO: p.parseResourceTag()
	}
	for !(p.atEnd() || (p.currentToken.IsSpecial() && IncludesInString("])}", p.currentToken.(scanner.ValueTokenInterface).ValueOfToken()))) {
		if returnFlag {
			panic("End of statement list encountered")
		}
		if p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == `^` {
			//TODO: smalltalk return statement ^
			panic("we should not have smalltalk return statement eg ^ in our script")
		} else {
			node := p.parseAssignment()
			statements = append(statements, node)
		}
		if p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == `.` {
			periods = append(periods, p.currentToken.GetStart())
			p.step()
			//TODO: comments
		} else {
			returnFlag = true
		}
		if p.emptyStatements {
			for p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == `.` {
				periods = append(periods, p.currentToken.GetStart())
				p.step()
			}
		}
	}
	sequenceNode.SetStatements(statements)
	sequenceNode.SetPeriods(periods)
	return sequenceNode
}

func (p *Parser) parseAssignment() treeNodes.ValueNodeInterface {
	if !(p.currentToken.IsIdentifier() && p.nextToken().IsAssignment()) {
		return p.parseCascadeMessage()
	}
	node := p.parseVariableNode()
	position := p.currentToken.GetStart()
	p.step()
	assignmentNode := treeNodes.NewAssignmentNode()
	assignmentNode.SetVariable(node)
	assignmentNode.SetValue(p.parseAssignment())
	assignmentNode.SetPosition(position)
	return assignmentNode
}

func (p *Parser) parseCascadeMessage() treeNodes.ValueNodeInterface {
	node := p.parseKeywordMessage()
	if !(p.currentToken.IsSpecial() && (p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ";" && node.IsMessage())) {
		return node
	}
	receiver := node.(treeNodes.NodeWithRreceiverInterface).GetReceiver()
	var messages []*treeNodes.MessageNode
	var semicolons []int64
	messages = append(messages, node.(*treeNodes.MessageNode))
	for p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ";" {
		semicolons = append(semicolons, p.currentToken.GetStart())
		p.step()
		var message *treeNodes.MessageNode
		if p.currentToken.IsIdentifier() {
			message = p.parseKeywordMessageWith(receiver).(*treeNodes.MessageNode)
		} else {
			if p.currentToken.IsLiteralToken() {
				p.patchNegativeLiteral()
			}
			if !p.currentToken.IsBinary() {
				panic("message expected")
			}
			temp := p.parseBinaryMessageWith(receiver)
			if temp == receiver {
				panic("message expected")
			}
			message = temp
		}
		messages = append(messages, message)
	}

	cascadeNode := treeNodes.NewCascadeNode()
	cascadeNode.SetSemicolons(semicolons)
	cascadeNode.SetMessages(messages)
	return cascadeNode
}

func (p *Parser) patchNegativeLiteral() {
	if !(p.currentToken.TypeOfToken() == scanner.NUMBER) {
		return
	}
	strVal := p.currentToken.(*scanner.NumberLiteralToken).ValueOfToken()
	value, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return
	}

	if value >= 0 {
		return
	}
	p.peekToken = p.currentToken
	p.currentToken = scanner.NewBinarySelectorToken(p.peekToken.GetStart(), `-`)
	p.peekToken.(*scanner.NumberLiteralToken).SetValue(strconv.FormatFloat(value*-1, 'f', 2, 64))
	if p.peekToken.TypeOfToken() == scanner.NUMBER {
		//TODO: working with source code for token
	}
	p.peekToken.SetStart(p.peekToken.GetStart() + 1)
}

func (p *Parser) parseKeywordMessage() treeNodes.ValueNodeInterface {
	return p.parseKeywordMessageWith(p.parseBinaryMessage())
}

func (p *Parser) parseKeywordMessageWith(valueNode treeNodes.ValueNodeInterface) treeNodes.ValueNodeInterface {
	var keywords []scanner.ValueTokenInterface
	var arguments []treeNodes.ValueNodeInterface
	isKeyword := false
	for p.currentToken.IsKeyword() {
		keywords = append(keywords, p.currentToken.(scanner.ValueTokenInterface))
		p.step()
		arguments = append(arguments, p.parseBinaryMessage())
		isKeyword = true
	}
	if isKeyword {
		node := treeNodes.NewMessageNode()
		node.SetReceiverSelectorPartsArguments(valueNode, keywords, arguments)
		return node
	} else {
		return valueNode
	}
}

func (p *Parser) parseBinaryMessage() treeNodes.ValueNodeInterface {
	node := p.parseUnaryMessage()
	for p.isBinaryAfterPatch() {
		node = p.parseBinaryMessageWith(node)
	}
	return node
}

func (p *Parser) isBinaryAfterPatch() bool {
	if p.currentToken.IsLiteralToken() {
		p.patchNegativeLiteral()
	}
	return p.currentToken.IsBinary()
}

func (p *Parser) parseBinaryMessageWith(nodeInterface treeNodes.ValueNodeInterface) *treeNodes.MessageNode {
	selector := p.currentToken.(*scanner.BinarySelectorToken)
	p.step()
	node := treeNodes.NewMessageNode()
	selectorParts := []scanner.ValueTokenInterface{selector} //literal array with one selector
	arguments := []treeNodes.ValueNodeInterface{p.parseUnaryMessage()}
	node.SetReceiverSelectorPartsArguments(nodeInterface, selectorParts, arguments)
	return node
}

func (p *Parser) parseUnaryMessage() treeNodes.ValueNodeInterface {
	node := p.parsePrimitiveObject()
	for p.currentToken.IsIdentifier() {
		//TODO: patchLiteralMessage
		node = p.parseUnaryMessageWith(node)
	}
	return node
}

func (p *Parser) parseUnaryMessageWith(nodeInterface treeNodes.ValueNodeInterface) *treeNodes.MessageNode {
	selector := p.currentToken.(*scanner.IdentifierToken)
	p.step()
	node := treeNodes.NewMessageNode()
	selectorParts := []scanner.ValueTokenInterface{selector} //literal array with one selector
	arguments := []treeNodes.ValueNodeInterface{}
	node.SetReceiverSelectorPartsArguments(nodeInterface, selectorParts, arguments)
	return node
}

func (p *Parser) parsePrimitiveObject() treeNodes.ValueNodeInterface {
	if p.currentToken.IsIdentifier() {
		return p.parsePrimitiveIdentifier()
	}
	if p.currentToken.IsLiteralToken() && !(p.currentToken.(scanner.LiteralTokenInterface).IsMultiKeyword()) {
		return p.parsePrimitiveLiteral()
	}
	if p.currentToken.IsLiteralArrayToken() {
		//TODO: ByteArray
		return p.parseLiteralArray()
	}
	if p.currentToken.IsSpecial() {
		if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "[" {
			return p.parseBlock()
		}
		if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "(" {
			return p.parseParenthesizedExpression()
		}
	}
	//in case of emergency LUL
	panic("what is our token?")
}

func (p *Parser) parseBlock() *treeNodes.BlockNode {
	position := p.currentToken.GetStart()
	p.step()
	node := treeNodes.NewBlockNode()
	p.parseBlockArgsInto(node)
	node.SetLeft(position)
	node.SetBody(p.parseStatements(false))
	if !(p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "]") {
		panic("Close bracket expected smth like ]")
	}
	node.SetRight(p.currentToken.GetStart())
	p.step()

	return node
}

func (p *Parser) parseBlockArgsInto(node *treeNodes.BlockNode) *treeNodes.BlockNode {
	var args []*treeNodes.VariableNode
	var colons []int64
	verticalBar := false
	for p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ":" {
		colons = append(colons, p.currentToken.GetStart())
		p.step()
		verticalBar = true
		args = append(args, p.parseVariableNode())
	}
	if verticalBar {
		if p.currentToken.IsBinary() {
			node.SetBar(p.currentToken.GetStart())
			if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "|" {
				p.step()
			} else {
				panic("bar inside block node is expected")
			}
		} else {
			if !(p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "]") {
				panic("bar inside block node is expected")
			}
		}
	}
	node.SetArguments(args)
	node.SetColons(colons)
	return node
}

func (p *Parser) parseParenthesizedExpression() treeNodes.ValueNodeInterface {
	leftParen := p.currentToken.GetStart()
	p.step()
	node := p.parseAssignment()
	if p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ")" {
		interval := treeNodes.Interval{}
		interval.SetStart(p.currentToken.GetStart())
		interval.SetStop(leftParen)
		node.AddParenthesis(interval)
		p.step()
		return node
	} else {
		panic("close parenthesis expected. something like ) ")
	}
}

func (p *Parser) parseLiteralArray() treeNodes.LiteralNodeInterface {
	var contents []treeNodes.LiteralNodeInterface
	start := p.currentToken.GetStart()
	p.step()
	for !(p.atEnd() || (p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ")")) {
		contents = append(contents, p.parseLiteralArrayObject())
	}
	if !(p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ")") {
		panic("hmm parse error btw. we expect ) here")
	}
	stop := p.currentToken.(scanner.ValueTokenInterface).GetStop()
	p.step()
	node := treeNodes.CreateLiteralArrayNode(start, stop, contents)
	return node
}

func (p *Parser) parseLiteralArrayObject() treeNodes.LiteralNodeInterface {
	if p.currentToken.IsSpecial() {
		if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "(" {
			return p.parseLiteralArray()
		}
		//TODO: ByteArray
		/*if p.currentToken.ValueOfToken() == "[" {
			return p.parseLiteralByteArray()
		}*/
	}
	if p.currentToken.IsLiteralArrayToken() {
		if p.currentToken.IsForByteArray() {
			//TODO: ByteArray
			return nil
		} else {
			return p.parseLiteralArray()
		}
	}
	//TODO: Optimized token
	//TODO: patchLiteralArrayToken
	return p.parsePrimitiveLiteral()
}

func (p *Parser) parsePrimitiveIdentifier() *treeNodes.VariableNode {
	token := p.currentToken.(scanner.ValueTokenInterface)
	p.step()
	node := &treeNodes.VariableNode{treeNodes.NewValueNode(), token}
	return node
}

func (p *Parser) parsePrimitiveLiteral() treeNodes.LiteralNodeInterface {
	token := p.currentToken.(scanner.LiteralTokenInterface)
	p.step()
	literalNode := treeNodes.NewLiteralNode()
	node := literalNode.LiteralToken(token)

	return node
}

func (p *Parser) step() {
	if p.peekToken != nil {
		p.currentToken = p.peekToken
		p.peekToken = nil
	} else {
		p.currentToken = p.scanner.Next()
	}
}

func (p *Parser) nextToken() scanner.TokenInterface {
	if p.peekToken == nil {
		p.peekToken = p.scanner.Next()
	}
	return p.peekToken
}

func (p *Parser) atEnd() bool {
	return p.currentToken.TypeOfToken() == "EOFToken"
}

func IncludesInString(arrayOfElements string, baseString string) bool {
	for _, char := range arrayOfElements {
		if strings.ContainsRune(baseString, char) {
			return true
		}
	}
	return false
}
