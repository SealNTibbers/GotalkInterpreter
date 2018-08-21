package scanner

import (
	"GotalkInterpreter"
	"GotalkInterpreter/testutils"
	"testing"
)

func TestScanNumber(t *testing.T) {
	inputString := `0.56`
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	token = vwScanner.Next()
	testutils.ASSERT_STREQ(t, NUMBER, token.TypeOfToken())
	testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), "0.56")

	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")
}

func TestScanIdentifier(t *testing.T) {
	inputString := `radio_altitude`
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	token = vwScanner.Next()
	testutils.ASSERT_STREQ(t, IDENT, token.TypeOfToken())
	testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), "radio_altitude")

	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")
}

func TestScanFloatNumberWithD(t *testing.T) {
	inputString := `1.02d`
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	token = vwScanner.Next()
	testutils.ASSERT_STREQ(t, NUMBER, token.TypeOfToken())
	testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), "1.02")

	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, "EOFToken", eofToken.TypeOfToken())
}

func TestScanFloatNumberWithExp(t *testing.T) {
	inputString := `1e4`
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	token = vwScanner.Next()
	testutils.ASSERT_STREQ(t, NUMBER, token.TypeOfToken())
	testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), "10000")

	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, "EOFToken", eofToken.TypeOfToken())
}

func TestScanFloatNumberWithNegativeExp(t *testing.T) {
	inputString := `1e-4`
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	token = vwScanner.Next()
	testutils.ASSERT_STREQ(t, NUMBER, token.TypeOfToken())
	testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), "0.0001")

	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")
}

func TestScanRealExpressions(t *testing.T) {
	inputString := `ikvsp_iaspeed_kmph\\10/10-0.9*10`

	tests := []struct {
		expectedTokenType string
		expectedValue     string
	}{
		{IDENT, "ikvsp_iaspeed_kmph"},
		{BIN, `\\`},
		{NUMBER, "10"},
		{BIN, `/`},
		{NUMBER, "10"},
		{NUMBER, "-0.9"},
		{BIN, "*"},
		{NUMBER, "10"},
	}
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	for _, eachTest := range tests {
		token = vwScanner.Next()
		testutils.ASSERT_STREQ(t, token.TypeOfToken(), eachTest.expectedTokenType)
		testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), eachTest.expectedValue)
	}
	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")
}

func TestScanIfStatementExpressionWithDifferentSubexpressions(t *testing.T) {
	inputString := `(abc > -137.74 abs) not ifTrue:['b'] ifFalse:[true]`

	tests := []struct {
		expectedTokenType string
		expectedValue     string
	}{
		{SPEC, "("},
		{IDENT, "abc"},
		{BIN, ">"},
		{NUMBER, "-137.74"},
		{IDENT, "abs"},
		{SPEC, ")"},
		{IDENT, "not"},
		{KEYWORD, "ifTrue:"},
		{SPEC, "["},
		{STRING, "b"},
		{SPEC, "]"},
		{KEYWORD, "ifFalse:"},
		{SPEC, "["},
		{BOOLEAN, "true"},
		{SPEC, "]"},
	}
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	for _, eachTest := range tests {
		token = vwScanner.Next()
		testutils.ASSERT_STREQ(t, token.TypeOfToken(), eachTest.expectedTokenType)
		testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), eachTest.expectedValue)
	}
	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")
}

func TestScanAssignmentExpression(t *testing.T) {
	inputString := `a:=1`

	tests := []struct {
		expectedTokenType string
		expectedValue     string
	}{
		{IDENT, "a"},
		{NUMBER, "1"},
	}
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	foundAssignment := false
	var index int
	for i := 0; i < len(tests)+1; i++ {
		token = vwScanner.Next()
		if foundAssignment {
			index = i - 1
		} else {
			index = i
		}

		if !token.IsAssignment() {
			testutils.ASSERT_STREQ(t, token.TypeOfToken(), tests[index].expectedTokenType)
			testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), tests[index].expectedValue)
		} else {
			foundAssignment = true
		}
	}
	testutils.ASSERT_TRUE(t, foundAssignment)

	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")
}

func TestScanArray(t *testing.T) {
	inputString := `#(1)`

	tests := []struct {
		expectedTokenType string
		expectedValue     string
	}{
		{ARRAY, "#("},
		{NUMBER, "1"},
		{SPEC, ")"},
	}
	vwReader := slysmalltalkinterpreter.NewReader(inputString)
	vwScanner := New(*vwReader)
	var token TokenInterface
	for _, eachTest := range tests {
		token = vwScanner.Next()
		testutils.ASSERT_STREQ(t, token.TypeOfToken(), eachTest.expectedTokenType)
		testutils.ASSERT_STREQ(t, token.(ValueTokenInterface).ValueOfToken(), eachTest.expectedValue)
	}
	eofToken := vwScanner.Next()
	testutils.ASSERT_STREQ(t, eofToken.TypeOfToken(), "EOFToken")

}
