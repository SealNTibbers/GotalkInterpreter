package lineedit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	left  = "\x1b[D"
	right = "\x1b[C"
	up    = "\x1b[A"
	down  = "\x1b[B"
	home  = "\x1b[H"
	end   = "\x1b[F"
	del   = "\x1b[3~"
)

// typeLines feeds keys to the editing loop and returns the lines it produced.
func typeLines(t *testing.T, history []string, keys string) []string {
	t.Helper()
	var out bytes.Buffer
	e := newEditor(strings.NewReader(keys), &out)
	for _, line := range history {
		e.History.Add(line)
	}
	var lines []string
	for {
		line, err := e.edit("st> ")
		if err == io.EOF {
			return lines
		}
		if err == ErrInterrupted {
			lines = append(lines, "<interrupted>")
			continue
		}
		lines = append(lines, line)
		e.History.Add(line)
	}
}

func TestEditing(t *testing.T) {
	cases := []struct{ keys, want string }{
		{"3 + 4\r", "3 + 4"},
		{"3 4" + left + left + "+\r", "3+ 4"},
		{"34" + left + " + " + right + "\r", "3 + 4"},
		{"bc" + home + "a" + end + "d\r", "abcd"},
		{"bc\x01a\x05d\r", "abcd"}, // Ctrl+A, Ctrl+E
		{"abc\x7f\x7fx\r", "ax"},
		{"abc" + home + del + "\r", "bc"},
		{"abc" + left + "\x04\r", "ab"},                      // Ctrl+D deletes under the cursor
		{"x := 1 + 2" + left + left + "\x0b\r", "x := 1 +"},  // Ctrl+K
		{"x := 1 + 2" + left + "\x15\r", "2"},                // Ctrl+U
		{"foo bar baz\x17\r", "foo bar "},                    // Ctrl+W
		{"foo bar" + "\x1bb" + "X\r", "foo Xbar"},            // Alt+B
		{"foo bar" + home + "\x1b[1;5C" + "X\r", "fooX bar"}, // Ctrl+Right
		{"héllo" + left + left + left + "\x7f" + "e\r", "hello"},
		{"ab" + left + left + left + right + right + right + right + "c\r", "abc"}, // stops at the ends
	}
	for _, c := range cases {
		got := typeLines(t, nil, c.keys)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("keys %q: got %q, want %q", c.keys, got, c.want)
		}
	}
}

func TestHistoryNavigation(t *testing.T) {
	past := []string{"first", "second", "third"}
	cases := []struct{ keys, want string }{
		{up + "\r", "third"},
		{up + up + up + up + "\r", "first"}, // stops at the oldest
		{up + up + down + "\r", "third"},
		{"draft" + up + down + "\r", "draft"}, // Down past the newest brings back what was typed
		{up + " + 1\r", "third + 1"},
		{"\x10\x10\r", "second"},              // Ctrl+P
		{"\x12sec\r", "second"},               // Ctrl+R, Enter runs the match
		{"\x12ir\x12\r", "first"},             // Ctrl+R again finds an older match
		{"\x12thi" + right + "!\r", "third!"}, // an arrow key leaves the search to edit the match
		{"\x12zzz\x07\r", ""},                 // Ctrl+G cancels
	}
	for _, c := range cases {
		got := typeLines(t, past, c.keys)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("keys %q: got %q, want %q", c.keys, got, c.want)
		}
	}
	// a new entry is recalled first
	if got := typeLines(t, past, "fourth\r"+up+"\r"); strings.Join(got, ",") != "fourth,fourth" {
		t.Errorf("got %q", got)
	}
}

func TestInterruptAndEOF(t *testing.T) {
	got := typeLines(t, nil, "half typed\x03done\r\x04")
	if strings.Join(got, ",") != "<interrupted>,done" {
		t.Errorf("got %q", got)
	}
}

func TestLongLineScrolls(t *testing.T) {
	var out bytes.Buffer
	e := newEditor(strings.NewReader(""), &out)
	e.width = func() int { return 20 }
	e.prompt = "st> "
	e.setLine(strings.Repeat("a", 30) + "XYZ")
	e.refresh()
	last := out.String()[strings.LastIndex(out.String(), "\r"+e.prompt):]
	if !strings.Contains(last, "XYZ") || strings.Contains(last, strings.Repeat("a", 16)) {
		t.Errorf("the end of a long line should be shown: %q", last)
	}
}

func TestHistoryFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history")
	h := LoadHistory(path, 3)
	for _, line := range []string{"a", "b", "b", "  ", "two\nlines", "c", "d"} {
		h.Add(line)
	}
	if got := strings.Join(h.Entries(), ","); got != "b,c,d" {
		t.Errorf("entries %q", got)
	}
	if got := strings.Join(LoadHistory(path, 3).Entries(), ","); got != "b,c,d" {
		t.Errorf("reloaded %q", got)
	}
	// a file far past the limit is compacted
	for i := 0; i < 10; i++ {
		h.Add(strings.Repeat("x", i+1))
	}
	LoadHistory(path, 3)
	data, _ := os.ReadFile(path)
	if lines := strings.Count(string(data), "\n"); lines != 3 {
		t.Errorf("file has %d lines after compaction", lines)
	}
}
