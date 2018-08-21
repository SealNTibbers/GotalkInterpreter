package evaluator

import (
	"github.com/SealNTibbers/GotalkInterpreter/testutils"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
	"testing"
)

func TestArrayEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	inputString = `#(1 2 3)`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.ARRAY_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(0).(*treeNodes.SmalltalkNumber).GetValue(), 1)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(1).(*treeNodes.SmalltalkNumber).GetValue(), 2)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(2).(*treeNodes.SmalltalkNumber).GetValue(), 3)

	inputString = `#(1 2 3)*2+4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.ARRAY_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(0).(*treeNodes.SmalltalkNumber).GetValue(), 6)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(1).(*treeNodes.SmalltalkNumber).GetValue(), 8)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(2).(*treeNodes.SmalltalkNumber).GetValue(), 10)

	inputString = `#(1 2 3) at:1`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 1)

	inputString = `#(#(1 2) #(3 4)) at:1`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.ARRAY_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(0).Value().(*treeNodes.SmalltalkNumber).GetValue(), 1)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(1).Value().(*treeNodes.SmalltalkNumber).GetValue(), 2)
}

func TestArrayMathEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	inputString = `#(1 2) * -3 + 4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.ARRAY_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(0).(*treeNodes.SmalltalkNumber).GetValue(), 1)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkArray).GetValueAt(1).(*treeNodes.SmalltalkNumber).GetValue(), -2)
}

func TestGlobalScopeEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	globalScope := new(treeNodes.Scope).Initialize()
	globalScope.SetVar("x", treeNodes.NewSmalltalkNumber(25))
	globalScope.SetVar("radio_altitude", treeNodes.NewSmalltalkNumber(25))

	inputString = `x+75`
	resultObject = TestEvalWithScope(inputString, globalScope)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 100)

	inputString = `[:v| v + x] value: 5`
	resultObject = TestEvalWithScope(inputString, globalScope)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 30)

	inputString = `radio_altitude`
	resultObject = TestEvalWithScope(inputString, globalScope)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 25)
}

func TestRealWorldEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	globalScope := new(treeNodes.Scope).Initialize()
	globalScope.SetVar("speed", treeNodes.NewSmalltalkNumber(25))
	globalScope.SetVar("angle", treeNodes.NewSmalltalkNumber(25))

	inputString = `angle\\10/10-0.9*10`
	resultObject = TestEvalWithScope(inputString, globalScope)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -4)

	inputString = `(((((-34.5+(speed*3.76)) degreesToRadians cos)*162)*(((-34.5+(speed*3.76)) degreesToRadians cos)*162)+106981)  sqrt - (((-34.5+(speed*3.76)) degreesToRadians cos)*162)) negated`
	resultObject = TestEvalWithScope(inputString, globalScope)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -255.0, 0.1)
}

func TestComplexBlockEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	inputString = `[5 + 7] value`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 12)

	inputString = `[:v| v + 7] value: 5`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 12)
}

func TestTemporariesEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	inputString = `|x| x := 5`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 5)

	inputString = `|x| x := -5. x abs`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 5)

	inputString = `|x| x := true. x ifTrue:[5] ifFalse: [0]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 5)
}

func TestBasicBooleanMessageEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	inputString = `true not`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `(5 < 1) not`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true and: [false]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true and: [true]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false and: [true]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false and: [false]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true or: [false]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true or: [true]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false or: [true]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false or: [false]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true xor: true`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false xor: true`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false xor: false`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true & true`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false & true`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false & false`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true | false`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `true | true`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `false | true`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
}

func TestIfTrueIfFalseStatementMessageEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface

	inputString = `true ifTrue:[5]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 5)

	inputString = `5 < 1 ifFalse:[false]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())

	inputString = `15 < 3 ifTrue:[7.45 - 0.45] ifFalse:[8 // 3]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 2)

	inputString = `15 < 3 ifFalse:[7.45 - 0.45] ifTrue:[8 // 3]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 7)

	inputString = `15 > 3 ifTrue:[(7.45 - 0.45) > 10 ifFalse:[32] ifTrue:[21]] ifFalse:[8 // 3]`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 32)
}

func TestMessageOrderingEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface
	inputString = `2 + 2 * 3`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 12)
	inputString = `2 + 3 max: 2`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 5)
	inputString = `3 - 5 abs`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -2)
}

func TestNumberMessageEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface
	inputString = `7.45 + 4.55`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 12)
	inputString = `7.45 - 0.45`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 7)
	inputString = `8 * 0.5`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 4)
	inputString = `8 / 0.5`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 16)
	inputString = `8 // 3`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 2)
	inputString = `8 \\ 4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 0)
	inputString = `9 rem: 4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 1)
	inputString = `8 max: 4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 8)
	inputString = `8 min: 4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 4)
	inputString = `-8 abs`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 8)
	inputString = `16 sqrt`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 4)
	inputString = `16 sqr`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 256)
	inputString = `30 sin`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -0.988, 0.001)
	inputString = `30 cos`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 0.15, 0.01)
	inputString = `30 tan`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -6.4, 0.01)
	inputString = `0.5 arcSin`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 0.52, 0.01)
	inputString = `0.5 arcCos`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 1.04, 0.01)
	inputString = `0.5 arcTan`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_NEAR(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 0.46, 0.01)
	inputString = `3.5 rounded`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 4)
	inputString = `3.5 truncated`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 3)
	inputString = `3.5 floor`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 3)
	inputString = `3.5 fractionPart`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 0.5)
	inputString = `3.5 ceiling`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 4)
}

func TestBooleanBinaryMessageEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface
	inputString = `7 > 4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
	inputString = `4 < 7`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
	inputString = `7 >= 6.9`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
	inputString = `4 <= 4.1`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
	inputString = `4.12 = 4.12`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
	inputString = `4.12 ~= 4.119`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_TRUE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
}

func TestNumbersEvaluation(t *testing.T) {
	var inputString string
	var resultObject treeNodes.SmalltalkObjectInterface
	inputString = `5`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 5)
	inputString = `-5`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -5)
	inputString = `0.56`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 0.56)
	inputString = `-0.56`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), -0.56)
	inputString = `1.2e4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 12000)
	inputString = `1.2e-4`
	resultObject = TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.NUMBER_OBJ)
	testutils.ASSERT_FLOAT64_EQ(t, resultObject.(*treeNodes.SmalltalkNumber).GetValue(), 1.2e-4)
}

func TestStringEvaluation(t *testing.T) {
	inputString := `'Smalltalk evaluator'`
	resultObject := TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.STRING_OBJ)
	testutils.ASSERT_STREQ(t, resultObject.(*treeNodes.SmalltalkString).GetValue(), "Smalltalk evaluator")
}

func TestBooleanEvaluation(t *testing.T) {
	inputString := `false`
	resultObject := TestEval(inputString)
	testutils.ASSERT_TRUE(t, resultObject.TypeOf() == treeNodes.BOOLEAN_OBJ)
	testutils.ASSERT_FALSE(t, resultObject.(*treeNodes.SmalltalkBoolean).GetValue())
}
