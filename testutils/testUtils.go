package testutils

import "testing"

func ASSERT_STREQ(t *testing.T, actual string, expected string) {
	if expected != actual {
		t.Fail()
		t.Fatalf("expected=%s, got=%s", expected, actual)
	}
}

func ASSERT_EQ(t *testing.T, actual int, expected int) {
	if expected != actual {
		t.Fatalf("expected=%d, got=%d", expected, actual)
	}
}

func ASSERT_FLOAT32_EQ(t *testing.T, actual float32, expected float32) {
	if expected != actual {
		t.Fatalf("expected=%e, got=%e", expected, actual)
	}
}

func ASSERT_FLOAT64_EQ(t *testing.T, actual float64, expected float64) {
	if expected != actual {
		t.Fatalf("expected=%e, got=%e", expected, actual)
	}
}

func ASSERT_NEAR(t *testing.T, actual float64, expected float64, abs_error float64) {
	if !((expected-actual) < abs_error && (actual-expected) < abs_error) {
		t.Fatalf("expected=%e, got=%e", expected, actual)
	}
}

func ASSERT_TRUE(t *testing.T, condition bool) {
	if !condition {
		t.Fatalf("We got False, but want True.")
	}
}

func ASSERT_FALSE(t *testing.T, condition bool) {
	if condition {
		t.Fatalf("We got True, but want False.")
	}
}
