package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNeedsMoreInput(t *testing.T) {
	for code, want := range map[string]bool{
		`3 + 4`:                  false,
		`3 +`:                    true,
		`x :=`:                   true,
		`a at: 1 put:`:           true,
		`[:x | x`:                true,
		`(1 + 2`:                 true,
		`#(1 2`:                  true,
		`'it''s`:                 true,
		`'it''s'`:                false,
		`"comment`:               true,
		`$( size`:                false, // a character literal is not a bracket
		`$' size`:                false,
		`3 ]`:                    false, // the parser reports it
		"[:x |\n x * 2]":         false,
		`foo bar`:                false, // complete; fails when evaluated
		"'a' ,\n":                true,
		"\"note\" 3 printString": false,
	} {
		if got := needsMoreInput(code); got != want {
			t.Errorf("needsMoreInput(%q) = %v, want %v", code, got, want)
		}
	}
}

func TestSplitChunks(t *testing.T) {
	chunks := splitChunks("a := 1!\n\n'x!!y' size! \"!\" b\n!")
	var codes []string
	var lines []int
	for _, c := range chunks {
		codes = append(codes, strings.TrimSpace(c.code))
		lines = append(lines, c.line)
	}
	want := []string{"a := 1", "'x!y' size", `"!" b`, ""}
	if strings.Join(codes, "|") != strings.Join(want, "|") {
		t.Errorf("chunks %q, want %q", codes, want)
	}
	if lines[0] != 1 || lines[1] != 3 || lines[2] != 3 {
		t.Errorf("chunk lines %v", lines)
	}
}

func session(input string) (string, string) {
	var out, errOut bytes.Buffer
	newConsole(&out, &errOut).repl(strings.NewReader(input))
	return out.String(), errOut.String()
}

func TestREPL(t *testing.T) {
	out, errOut := session(strings.Join([]string{
		`x := 3 + 4.`,
		`x * 2`,
		`[:a :b |`,
		`  a + b] value: x value: 1`,
		`"only a comment"`,
		`Transcript show: 'Hi'; show: ' there'`,
		`Transcript print: 6 * 7; cr`,
		`'two`,
		`lines' size`,
		`foo bar`,
		`:reset`,
		`x`,
		`:quit`,
		`99`,
	}, "\n"))
	for _, want := range []string{"st> 7\n", "st> 14\n", "  > 8\n", "Hi there\n", "42\n", "  > 9\n", "All variables forgotten."} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "nil") || strings.Contains(out, "a Transcript") || strings.Contains(out, "99") {
		t.Errorf("unexpected output:\n%s", out)
	}
	if !strings.Contains(errOut, "undefined variable: foo") || !strings.Contains(errOut, "undefined variable: x") {
		t.Errorf("errors:\n%s", errOut)
	}
}

func TestRunScript(t *testing.T) {
	var out, errOut bytes.Buffer
	c := newConsole(&out, &errOut)
	c.runReader("t.st", strings.NewReader("#!/usr/bin/env gotalk\nTranscript show: 'a'!\n\n  nope frob!\nTranscript show: 'b!!'\n"))
	if out.String() != "a\nb!\n" { // an error starts on a new line
		t.Errorf("output %q", out.String())
	}
	if errOut.String() != "t.st:4: Error: undefined variable: nope\n" {
		t.Errorf("errors %q", errOut.String())
	}
	if !c.failed {
		t.Error("a failed chunk must fail the run")
	}
}
