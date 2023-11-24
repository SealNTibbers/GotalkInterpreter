package parser

import (
	"errors"
	"strconv"
	"strings"

	"github.com/SealNTibbers/GotalkInterpreter/scanner"
	"github.com/SealNTibbers/GotalkInterpreter/talkio"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

type Parser struct {
	scanner         *scanner.Scanner
	currentToken    scanner.TokenInterface
	peekToken       scanner.TokenInterface
	emptyStatements bool
}

func InitializeParserFor(expressionString string) (treeNodes.ProgramNodeInterface, error) {
	reader := talkio.NewReader(expressionString)
	scanner := scanner.New(*reader)
	parser := &Parser{scanner, nil, nil, false}

	//initialize struct members
	err := parser.step()
	if err != nil {
		return nil, err
	}
	node, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
	if len(node.GetStatements()) == 1 && len(node.GetTemporaries()) == 0 {
		return node.GetStatements()[0], nil
	} else {
		return node, nil
	}
}

func (p *Parser) parseExpression() (*treeNodes.SequenceNode, error) {
	return p.parseStatements(false)
}

func (p *Parser) parseStatements(tagBool bool) (*treeNodes.SequenceNode, error) {
	var leftBar int64
	var rightBar int64
	var args []*treeNodes.VariableNode
	var err error
	if p.currentToken.IsBinary() {
		if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "|" {
			leftBar = p.currentToken.GetStart()
			err = p.step()
			if err != nil {
				return nil, err
			}
			args, err = p.parseArgs()
			if err != nil || !(p.currentToken.IsBinary() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "|") {
				return nil, errors.New("Parse error in parseStatements function.")
			}
			rightBar = p.currentToken.GetStart()
			err = p.step()
			if err != nil {
				return nil, err
			}
		} else {
			if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "||" {
				leftBar = p.currentToken.GetStart()
				rightBar = leftBar + 1
				err = p.step()
				if err != nil {
					return nil, err
				}
			}
		}
	}
	node := treeNodes.NewSequenceNode()
	node.SetLeftBar(leftBar)
	node.SetRightBar(rightBar)
	node.SetTemporaries(args)
	parsedStatementList, err := p.parseStatementListInto(tagBool, node)
	if err != nil {
		return nil, err
	}
	return parsedStatementList, nil
}

func (p *Parser) parseArgs() ([]*treeNodes.VariableNode, error) {
	var args []*treeNodes.VariableNode
	for p.currentToken.IsIdentifier() {
		parsedVar, err := p.parseVariableNode()
		if err != nil {
			return nil, err
		}
		args = append(args, parsedVar)
	}
	return args, nil
}

func (p *Parser) parseVariableNode() (*treeNodes.VariableNode, error) {
	if p.currentToken.IsIdentifier() {
		return p.parsePrimitiveIdentifier()
	} else {
		return nil, errors.New("we expect variable name here btw")
	}
}

func (p *Parser) parseStatementListInto(tagBool bool, sequenceNode *treeNodes.SequenceNode) (*treeNodes.SequenceNode, error) {
	returnFlag := false
	var periods []int64
	var statements []treeNodes.ProgramNodeInterface
	if tagBool {
		//TODO: p.parseResourceTag()
	}
	for !(p.atEnd() || (p.currentToken.IsSpecial() && IncludesInString("])}", p.currentToken.(scanner.ValueTokenInterface).ValueOfToken()))) {
		if returnFlag {
			return nil, errors.New("End of statement list encountered")
		}
		if p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == `^` {
			//TODO: smalltalk return statement ^
			return nil, errors.New("we should not have smalltalk return statement eg ^ in our script")
		} else {
			node, err := p.parseAssignment()
			if err != nil {
				return nil, err
			}
			statements = append(statements, node)
		}
		if p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == `.` {
			periods = append(periods, p.currentToken.GetStart())
			err := p.step()
			if err != nil {
				return nil, err
			}
			//TODO: comments
		} else {
			returnFlag = true
		}
		if p.emptyStatements {
			for p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == `.` {
				periods = append(periods, p.currentToken.GetStart())
				err := p.step()
				if err != nil {
					return nil, err
				}
			}
		}
	}
	sequenceNode.SetStatements(statements)
	sequenceNode.SetPeriods(periods)
	return sequenceNode, nil
}

func (p *Parser) parseAssignment() (treeNodes.ValueNodeInterface, error) {
	if !(p.currentToken.IsIdentifier() && p.nextToken().IsAssignment()) {
		return p.parseCascadeMessage()
	}
	node, err := p.parseVariableNode()
	if err != nil {
		return nil, err
	}
	position := p.currentToken.GetStart()
	err = p.step()
	if err != nil {
		return nil, err
	}
	assignmentNode := treeNodes.NewAssignmentNode()
	assignmentNode.SetVariable(node)
	parsedValue, err := p.parseAssignment()
	if err != nil {
		return nil, err
	}
	assignmentNode.SetValue(parsedValue)
	assignmentNode.SetPosition(position)
	return assignmentNode, nil
}

func (p *Parser) parseCascadeMessage() (treeNodes.ValueNodeInterface, error) {
	var err error
	node, err := p.parseKeywordMessage()
	if err != nil {
		return nil, err
	}
	if !(p.currentToken.IsSpecial() && (p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ";" && node.IsMessage())) {
		return node, nil
	}
	receiver := node.(treeNodes.NodeWithRreceiverInterface).GetReceiver()
	var messages []*treeNodes.MessageNode
	var semicolons []int64
	messages = append(messages, node.(*treeNodes.MessageNode))
	for p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ";" {
		semicolons = append(semicolons, p.currentToken.GetStart())
		err = p.step()
		if err != nil {
			return nil, err
		}
		var message *treeNodes.MessageNode
		if p.currentToken.IsIdentifier() {
			tmpMsg, err := p.parseKeywordMessageWith(receiver)
			if err != nil {
				return nil, err
			}
			message = tmpMsg.(*treeNodes.MessageNode)
		} else {
			if p.currentToken.IsLiteralToken() {
				p.patchNegativeLiteral()
			}
			if !p.currentToken.IsBinary() {
				return nil, errors.New("message expected")
			}
			temp, err := p.parseBinaryMessageWith(receiver)
			if err != nil {
				return nil, err
			}
			if temp == receiver {
				return nil, errors.New("message expected")
			}
			message = temp
		}
		messages = append(messages, message)
	}

	cascadeNode := treeNodes.NewCascadeNode()
	cascadeNode.SetSemicolons(semicolons)
	cascadeNode.SetMessages(messages)
	return cascadeNode, nil
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

func (p *Parser) parseKeywordMessage() (treeNodes.ValueNodeInterface, error) {
	parsedBinary, err := p.parseBinaryMessage()
	if err != nil {
		return nil, err
	}
	return p.parseKeywordMessageWith(parsedBinary)
}

func (p *Parser) parseKeywordMessageWith(valueNode treeNodes.ValueNodeInterface) (treeNodes.ValueNodeInterface, error) {
	var keywords []scanner.ValueTokenInterface
	var arguments []treeNodes.ValueNodeInterface
	isKeyword := false
	for p.currentToken.IsKeyword() {
		keywords = append(keywords, p.currentToken.(scanner.ValueTokenInterface))
		err := p.step()
		if err != nil {
			return nil, err
		}
		parsedBinary, err := p.parseBinaryMessage()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, parsedBinary)
		isKeyword = true
	}
	if isKeyword {
		node := treeNodes.NewMessageNode()
		node.SetReceiverSelectorPartsArguments(valueNode, keywords, arguments)
		return node, nil
	} else {
		return valueNode, nil
	}
}

func (p *Parser) parseBinaryMessage() (treeNodes.ValueNodeInterface, error) {
	node, err := p.parseUnaryMessage()
	if err != nil {
		return nil, err
	}
	for p.isBinaryAfterPatch() {
		node, err = p.parseBinaryMessageWith(node)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (p *Parser) isBinaryAfterPatch() bool {
	if p.currentToken.IsLiteralToken() {
		p.patchNegativeLiteral()
	}
	return p.currentToken.IsBinary()
}

func (p *Parser) parseBinaryMessageWith(nodeInterface treeNodes.ValueNodeInterface) (*treeNodes.MessageNode, error) {
	selector := p.currentToken.(*scanner.BinarySelectorToken)
	err := p.step()
	if err != nil {
		return nil, err
	}
	node := treeNodes.NewMessageNode()
	selectorParts := []scanner.ValueTokenInterface{selector}
	parsedUnary, err := p.parseUnaryMessage()
	if err != nil {
		return nil, err
	}
	//literal array with one selector
	arguments := []treeNodes.ValueNodeInterface{parsedUnary}
	node.SetReceiverSelectorPartsArguments(nodeInterface, selectorParts, arguments)
	return node, nil
}

func (p *Parser) parseUnaryMessage() (treeNodes.ValueNodeInterface, error) {
	node, err := p.parsePrimitiveObject()
	if err != nil {
		return nil, err
	}
	for p.currentToken.IsIdentifier() {
		//TODO: patchLiteralMessage
		node, err = p.parseUnaryMessageWith(node)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (p *Parser) parseUnaryMessageWith(nodeInterface treeNodes.ValueNodeInterface) (*treeNodes.MessageNode, error) {
	selector := p.currentToken.(*scanner.IdentifierToken)
	err := p.step()
	if err != nil {
		return nil, err
	}
	node := treeNodes.NewMessageNode()
	selectorParts := []scanner.ValueTokenInterface{selector} //literal array with one selector
	arguments := []treeNodes.ValueNodeInterface{}
	node.SetReceiverSelectorPartsArguments(nodeInterface, selectorParts, arguments)
	return node, nil
}

func (p *Parser) parsePrimitiveObject() (treeNodes.ValueNodeInterface, error) {
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
	return nil, errors.New("what is our token?")
}

func (p *Parser) parseBlock() (*treeNodes.BlockNode, error) {
	position := p.currentToken.GetStart()
	err := p.step()
	if err != nil {
		return nil, err
	}
	node := treeNodes.NewBlockNode()
	_, err = p.parseBlockArgsInto(node)
	if err != nil {
		return nil, err
	}
	node.SetLeft(position)
	parsedStatements, err := p.parseStatements(false)
	if err != nil {
		return nil, err
	}
	node.SetBody(parsedStatements)
	if !(p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "]") {
		return nil, errors.New("Close bracket expected smth like ]")
	}
	node.SetRight(p.currentToken.GetStart())
	err = p.step()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseBlockArgsInto(node *treeNodes.BlockNode) (*treeNodes.BlockNode, error) {
	var args []*treeNodes.VariableNode
	var colons []int64
	verticalBar := false
	for p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ":" {
		colons = append(colons, p.currentToken.GetStart())
		err := p.step()
		if err != nil {
			return nil, err
		}
		verticalBar = true
		parsedVariable, err := p.parseVariableNode()
		if err != nil {
			return nil, err
		}
		args = append(args, parsedVariable)
	}
	if verticalBar {
		if p.currentToken.IsBinary() {
			node.SetBar(p.currentToken.GetStart())
			if p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "|" {
				err := p.step()
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("bar inside block node is expected")
			}
		} else {
			if !(p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == "]") {
				return nil, errors.New("bar inside block node is expected")
			}
		}
	}
	node.SetArguments(args)
	node.SetColons(colons)
	return node, nil
}

func (p *Parser) parseParenthesizedExpression() (treeNodes.ValueNodeInterface, error) {
	leftParen := p.currentToken.GetStart()
	err := p.step()
	if err != nil {
		return nil, err
	}
	node, err := p.parseAssignment()
	if err != nil {
		return nil, err
	}
	if p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ")" {
		interval := treeNodes.Interval{}
		interval.SetStart(p.currentToken.GetStart())
		interval.SetStop(leftParen)
		node.AddParenthesis(interval)
		err = p.step()
		if err != nil {
			return nil, err
		}
		return node, nil
	} else {
		return nil, errors.New("close parenthesis expected. something like ) ")
	}
}

func (p *Parser) parseLiteralArray() (treeNodes.LiteralNodeInterface, error) {
	var contents []treeNodes.LiteralNodeInterface
	start := p.currentToken.GetStart()
	err := p.step()
	if err != nil {
		return nil, err
	}
	for !(p.atEnd() || (p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ")")) {
		parsedLiteralArray, err := p.parseLiteralArrayObject()
		if err != nil {
			return nil, err
		}
		contents = append(contents, parsedLiteralArray)
	}
	if !(p.currentToken.IsSpecial() && p.currentToken.(scanner.ValueTokenInterface).ValueOfToken() == ")") {
		return nil, errors.New("hmm parse error btw. we expect ) here")
	}
	stop := p.currentToken.(scanner.ValueTokenInterface).GetStop()
	err = p.step()
	if err != nil {
		return nil, err
	}
	node := treeNodes.CreateLiteralArrayNode(start, stop, contents)
	return node, nil
}

func (p *Parser) parseLiteralArrayObject() (treeNodes.LiteralNodeInterface, error) {
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
			return nil, errors.New("Not implemented")
		} else {
			return p.parseLiteralArray()
		}
	}
	//TODO: Optimized token
	//TODO: patchLiteralArrayToken
	return p.parsePrimitiveLiteral()
}

func (p *Parser) parsePrimitiveIdentifier() (*treeNodes.VariableNode, error) {
	token := p.currentToken.(scanner.ValueTokenInterface)
	err := p.step()
	if err != nil {
		return nil, err
	}
	node := &treeNodes.VariableNode{ValueNode: treeNodes.NewValueNode(), Token: token}
	return node, nil
}

func (p *Parser) parsePrimitiveLiteral() (treeNodes.LiteralNodeInterface, error) {
	token := p.currentToken.(scanner.LiteralTokenInterface)
	err := p.step()
	if err != nil {
		return nil, err
	}
	literalNode := treeNodes.NewLiteralNode()
	node := literalNode.LiteralToken(token)

	return node, nil
}

func (p *Parser) step() error {
	if p.peekToken != nil {
		p.currentToken = p.peekToken
		p.peekToken = nil
	} else {
		currentToken, err := p.scanner.Next()
		if err != nil {
			return err
		}
		p.currentToken = currentToken
	}
	return nil
}

func (p *Parser) nextToken() scanner.TokenInterface {
	if p.peekToken == nil {
		peekToken, _ := p.scanner.Next()
		p.peekToken = peekToken
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
