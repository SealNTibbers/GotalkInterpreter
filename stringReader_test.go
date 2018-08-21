package slysmalltalkinterpreter

import (
	"github.com/SealNTibbers/GotalkInterpreter/testutils"
	"io"
	"testing"
)

func TestReadRune(t *testing.T) {
	inputString := "a + b"
	stringReader := NewReader(inputString)
	if len(inputString) != stringReader.Len() {
		t.Fatalf("Lenght of reader content must be equal leght of input string.")
	}

	tests := []struct {
		expectedRune rune
		expectedSize int
	}{
		{'a', 1},
		{' ', 1},
		{'+', 1},
		{' ', 1},
		{'b', 1},
	}

	countOfTests := len(tests)
	// test normal way
	for i, eachTest := range tests {
		{
			testutils.ASSERT_TRUE(t, int64(i) == stringReader.GetPosition())
			if i < countOfTests-1 {
				peekRune, err := stringReader.PeekRuneError()
				testutils.ASSERT_TRUE(t, err == nil)
				expectedPeekRune := tests[i].expectedRune

				testutils.ASSERT_STREQ(t, string(expectedPeekRune), string(peekRune))
			}
			rune, size, err := stringReader.ReadRune()
			testutils.ASSERT_TRUE(t, err == nil)
			testutils.ASSERT_STREQ(t, string(eachTest.expectedRune), string(rune))
			testutils.ASSERT_EQ(t, size, eachTest.expectedSize)
		}
	}
	// test wrong way, when we have been standing outside the reading stream.
	rune, size, err := stringReader.ReadRune()
	testutils.ASSERT_TRUE(t, rune == 0)
	testutils.ASSERT_TRUE(t, size == 0)
	testutils.ASSERT_TRUE(t, err == io.EOF)
}
