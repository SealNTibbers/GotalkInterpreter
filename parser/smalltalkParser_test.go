package parser

import (
	"testing"

	"github.com/SealNTibbers/GotalkInterpreter/scanner"
	"github.com/SealNTibbers/GotalkInterpreter/testutils"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

func TestNumberParser(t *testing.T) {
	inputString := `5.1`
	literalNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, literalNode.(*treeNodes.LiteralValueNode).GetValue(), "5.1")
}

func TestStringParser(t *testing.T) {
	inputString := `'test'`
	identifierNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, identifierNode.(*treeNodes.LiteralValueNode).GetValue(), "test")
}

func TestBoolParser(t *testing.T) {
	inputString := `false`
	identifierNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, identifierNode.(*treeNodes.LiteralValueNode).GetValue(), "false")
}

func TestAssignmentNumberParser(t *testing.T) {
	inputString := `a := 5`
	assignmentNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetVariable().Token.ValueOfToken(), "a")
	testutils.ASSERT_TRUE(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().IsLiteralNode())
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().(*treeNodes.LiteralValueNode).GetValue(), "5")
}

func TestAssignmentStringParser(t *testing.T) {
	inputString := `a := 'b'`
	assignmentNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetVariable().Token.ValueOfToken(), "a")
	testutils.ASSERT_TRUE(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().IsLiteralNode())
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().(*treeNodes.LiteralValueNode).GetValue(), "b")
}

func TestAssignmentBoolParser(t *testing.T) {
	inputString := `a := true`
	assignmentNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetVariable().Token.ValueOfToken(), "a")
	testutils.ASSERT_TRUE(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().IsLiteralNode())
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().(*treeNodes.LiteralValueNode).GetValue(), "true")
}

func TestAssignmentLiteralArrayParser(t *testing.T) {
	inputString := `a := #(1 2)`
	assignmentNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetVariable().Token.ValueOfToken(), "a")
	testutils.ASSERT_TRUE(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().IsLiteralArray())
	testutils.ASSERT_STREQ(t, assignmentNode.(*treeNodes.AssignmentNode).GetValue().(*treeNodes.LiteralArrayNode).GetValue(), "1 2")
}

func TestBinaryMessageParser(t *testing.T) {
	inputString := `a + b`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "a")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "+")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.VariableNode).GetName(), "b")
}

func TestNumberWithMinusParser(t *testing.T) {
	inputString := `2*3-4`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_TRUE(t, messageNode.IsMessage())
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.LiteralValueNode).GetValue(), "2")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "*")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.LiteralValueNode).GetValue(), "3")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "-")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.LiteralValueNode).GetValue(), "4.00")
}

func TestFewBinaryMessageParser(t *testing.T) {
	inputString := `a - b + c`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_TRUE(t, messageNode.IsMessage())
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "a")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "-")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.VariableNode).GetName(), "b")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "+")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.VariableNode).GetName(), "c")
}

func TestGroupMessageParser(t *testing.T) {
	inputString := `a - (b + c)`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_TRUE(t, messageNode.IsMessage())
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "a")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "-")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_TRUE(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].IsMessage())
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "b")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "+")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.VariableNode).GetName(), "c")

}

func TestUnaryMessageParser(t *testing.T) {
	inputString := `-1 abs`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.LiteralValueNode).GetValue(), "-1")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.IdentifierToken).ValueOfToken(), "abs")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()) == 0)
}

func TestIfStatementParser(t *testing.T) {
	inputString := `a > 10 ifTrue:[25] ifFalse:[2]`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_TRUE(t, messageNode.(*treeNodes.MessageNode).GetReceiver().IsMessage())
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "a")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), ">")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.LiteralValueNode).GetValue(), "10")
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelector(), "ifTrue:ifFalse:")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetSelectorParts()) == 2)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[0].ValueOfToken(), "ifTrue:")
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetSelectorParts()[1].ValueOfToken(), "ifFalse:")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()) == 2)
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.BlockNode).GetBody().GetStatements()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.BlockNode).GetBody().GetStatements()[0].(*treeNodes.LiteralValueNode).GetValue(), "25")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetArguments()[1].(*treeNodes.BlockNode).GetBody().GetStatements()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[1].(*treeNodes.BlockNode).GetBody().GetStatements()[0].(*treeNodes.LiteralValueNode).GetValue(), "2")
}

func TestIfStatementWithExpressionParser(t *testing.T) {
	inputString := `(pitch>0.9 ifTrue:[1] ifFalse:[0])*23`
	messageNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_TRUE(t, messageNode.(*treeNodes.MessageNode).GetReceiver().IsMessage())
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "pitch")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), ">")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.LiteralValueNode).GetValue(), "0.9")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()) == 2)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()[0].ValueOfToken(), "ifTrue:")
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetSelectorParts()[1].ValueOfToken(), "ifFalse:")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()) == 2)
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.BlockNode).GetBody().GetStatements()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.BlockNode).GetBody().GetStatements()[0].(*treeNodes.LiteralValueNode).GetValue(), "1")
	testutils.ASSERT_TRUE(t, len(messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[1].(*treeNodes.BlockNode).GetBody().GetStatements()) == 1)
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetReceiver().(*treeNodes.MessageNode).GetArguments()[1].(*treeNodes.BlockNode).GetBody().GetStatements()[0].(*treeNodes.LiteralValueNode).GetValue(), "0")
	testutils.ASSERT_STREQ(t, messageNode.(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.LiteralValueNode).GetValue(), "23")
}

func TestSequenceParser(t *testing.T) {
	inputString := `a := 2. a + 2`
	sequenceNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_TRUE(t, len(sequenceNode.(*treeNodes.SequenceNode).GetStatements()) == 2)
	testutils.ASSERT_TRUE(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[0].IsAssignment())
	testutils.ASSERT_STREQ(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[0].(*treeNodes.AssignmentNode).GetVariable().Token.ValueOfToken(), "a")
	testutils.ASSERT_TRUE(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[0].(*treeNodes.AssignmentNode).GetValue().IsLiteralNode())
	testutils.ASSERT_STREQ(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[0].(*treeNodes.AssignmentNode).GetValue().(*treeNodes.LiteralValueNode).GetValue(), "2")
	testutils.ASSERT_TRUE(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[1].IsMessage())
	testutils.ASSERT_STREQ(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[1].(*treeNodes.MessageNode).GetReceiver().(*treeNodes.VariableNode).GetName(), "a")
	testutils.ASSERT_TRUE(t, len(sequenceNode.(*treeNodes.SequenceNode).GetStatements()[1].(*treeNodes.MessageNode).GetSelectorParts()) == 1)
	testutils.ASSERT_STREQ(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[1].(*treeNodes.MessageNode).GetSelectorParts()[0].(*scanner.BinarySelectorToken).ValueOfToken(), "+")
	testutils.ASSERT_TRUE(t, len(sequenceNode.(*treeNodes.SequenceNode).GetStatements()[1].(*treeNodes.MessageNode).GetArguments()) == 1)
	testutils.ASSERT_STREQ(t, sequenceNode.(*treeNodes.SequenceNode).GetStatements()[1].(*treeNodes.MessageNode).GetArguments()[0].(*treeNodes.LiteralValueNode).GetValue(), "2")
}

func TestIdentifierParser(t *testing.T) {
	inputString := `radio_altitude`
	variableNode, _ := InitializeParserFor(inputString)
	testutils.ASSERT_STREQ(t, variableNode.(*treeNodes.VariableNode).GetName(), "radio_altitude")
}
