// Package lineedit reads lines from a terminal with editing and history, like a small readline: arrow keys,
// Home/End, word moves, kill commands, Up/Down history and Ctrl+R search. It uses only the standard library.
// When the input is not a terminal (a pipe, a file, or an unsupported OS), it reads plain lines.
package lineedit

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// ErrInterrupted is returned by ReadLine when the user presses Ctrl+C.
var ErrInterrupted = errors.New("interrupted")

// Editor reads lines from a terminal.
type Editor struct {
	History *History

	in    *bufio.Reader
	out   io.Writer
	fd    int
	raw   bool // fd is a terminal that supports raw mode
	width func() int

	// state of the line being edited
	prompt string
	buf    []rune
	pos    int
	offset int // first rune shown when the line is longer than the terminal
}

// New returns an editor reading from in and echoing to out. The history is kept in memory only; set History to
// a LoadHistory result to keep it in a file.
func New(in *os.File, out io.Writer) *Editor {
	e := newEditor(in, out)
	e.fd = int(in.Fd())
	e.raw = isTerminal(e.fd)
	e.width = func() int { return terminalWidth(e.fd) }
	return e
}

func newEditor(in io.Reader, out io.Writer) *Editor {
	return &Editor{History: NewHistory(), in: bufio.NewReader(in), out: out, width: func() int { return 80 }}
}

// ReadLine shows prompt and returns the line the user typed, without the newline. It returns io.EOF on Ctrl+D
// on an empty line (or at the end of the input) and ErrInterrupted on Ctrl+C.
func (e *Editor) ReadLine(prompt string) (string, error) {
	if !e.raw {
		return e.readPlain(prompt)
	}
	restore, err := enableRawMode(e.fd)
	if err != nil {
		return e.readPlain(prompt)
	}
	defer restore()
	return e.edit(prompt)
}

func (e *Editor) readPlain(prompt string) (string, error) {
	fmt.Fprint(e.out, prompt)
	line, err := e.in.ReadString('\n')
	if line == "" && err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// keys
const (
	ctrlA     = 1
	ctrlB     = 2
	ctrlC     = 3
	ctrlD     = 4
	ctrlE     = 5
	ctrlF     = 6
	ctrlG     = 7
	ctrlH     = 8
	tab       = 9
	ctrlJ     = 10
	ctrlK     = 11
	ctrlL     = 12
	enter     = 13
	ctrlN     = 14
	ctrlP     = 16
	ctrlR     = 18
	ctrlT     = 20
	ctrlU     = 21
	ctrlW     = 23
	escape    = 27
	backspace = 127
)

// edit runs the editing loop on a terminal in raw mode.
func (e *Editor) edit(prompt string) (string, error) {
	e.prompt, e.buf, e.pos, e.offset = prompt, nil, 0, 0
	history := e.History.entries()
	index := len(history) // len(history) is the line being typed
	draft := ""
	recall := func(i int) {
		if index == len(history) {
			draft = string(e.buf)
		}
		index = i
		if i == len(history) {
			e.setLine(draft)
		} else {
			e.setLine(history[i])
		}
	}
	e.refresh()
	for {
		r, _, err := e.in.ReadRune()
		if err != nil {
			if len(e.buf) > 0 {
				e.finish()
				return string(e.buf), nil
			}
			return "", io.EOF
		}
		switch r {
		case enter, ctrlJ:
			e.finish()
			return string(e.buf), nil
		case ctrlC:
			e.pos = len(e.buf)
			e.refresh()
			fmt.Fprint(e.out, "^C\r\n")
			return "", ErrInterrupted
		case ctrlD:
			if len(e.buf) == 0 {
				fmt.Fprint(e.out, "\r\n")
				return "", io.EOF
			}
			e.deleteAt(e.pos)
		case backspace, ctrlH:
			if e.pos > 0 {
				e.pos--
				e.deleteAt(e.pos)
			}
		case ctrlA:
			e.pos = 0
		case ctrlE:
			e.pos = len(e.buf)
		case ctrlB:
			e.move(-1)
		case ctrlF:
			e.move(1)
		case ctrlP:
			if index > 0 {
				recall(index - 1)
			}
		case ctrlN:
			if index < len(history) {
				recall(index + 1)
			}
		case ctrlK:
			e.buf = e.buf[:e.pos]
		case ctrlU:
			e.buf = append([]rune{}, e.buf[e.pos:]...)
			e.pos = 0
		case ctrlW:
			start := e.wordStart()
			e.buf = append(e.buf[:start], e.buf[e.pos:]...)
			e.pos = start
		case ctrlT:
			if e.pos > 0 && len(e.buf) > 1 {
				if e.pos == len(e.buf) {
					e.pos--
				}
				e.buf[e.pos-1], e.buf[e.pos] = e.buf[e.pos], e.buf[e.pos-1]
				e.pos++
			}
		case ctrlL:
			fmt.Fprint(e.out, "\x1b[H\x1b[2J")
		case ctrlR:
			line, accepted, err := e.search(history)
			if err != nil {
				return "", err
			}
			e.prompt = prompt
			e.setLine(line)
			if accepted {
				e.finish()
				return line, nil
			}
		case tab:
			e.insert(' ', ' ')
		case escape:
			switch e.escapeSequence() {
			case "left":
				e.move(-1)
			case "right":
				e.move(1)
			case "up":
				if index > 0 {
					recall(index - 1)
				}
			case "down":
				if index < len(history) {
					recall(index + 1)
				}
			case "home":
				e.pos = 0
			case "end":
				e.pos = len(e.buf)
			case "delete":
				e.deleteAt(e.pos)
			case "wordLeft":
				e.pos = e.wordStart()
			case "wordRight":
				e.pos = e.wordEnd()
			case "deleteWord":
				end := e.wordEnd()
				e.buf = append(e.buf[:e.pos], e.buf[end:]...)
			}
		default:
			if unicode.IsPrint(r) {
				e.insert(r)
			}
		}
		e.refresh()
	}
}

// escapeSequence reads the rest of an escape sequence (arrow keys, Home, Alt+key, ...) and names it.
func (e *Editor) escapeSequence() string {
	r, _, err := e.in.ReadRune()
	if err != nil {
		return ""
	}
	switch r {
	case 'b', 'B':
		return "wordLeft"
	case 'f', 'F':
		return "wordRight"
	case 'd', 'D':
		return "deleteWord"
	case backspace:
		return "" // Alt+Backspace: ignored
	case '[', 'O':
	default:
		return ""
	}
	// CSI: parameters (digits and ;) then a final byte
	var params strings.Builder
	for {
		c, _, err := e.in.ReadRune()
		if err != nil {
			return ""
		}
		if c >= 0x40 && c <= 0x7e {
			return csiName(params.String(), c)
		}
		params.WriteRune(c)
	}
}

func csiName(params string, final rune) string {
	modified := strings.HasSuffix(params, ";5") || strings.HasSuffix(params, ";3") // Ctrl or Alt
	switch final {
	case 'A':
		return "up"
	case 'B':
		return "down"
	case 'C':
		if modified {
			return "wordRight"
		}
		return "right"
	case 'D':
		if modified {
			return "wordLeft"
		}
		return "left"
	case 'H':
		return "home"
	case 'F':
		return "end"
	case '~':
		switch params {
		case "1", "7":
			return "home"
		case "4", "8":
			return "end"
		case "3":
			return "delete"
		}
	}
	return ""
}

// search is Ctrl+R: an incremental search back through the history. It returns the chosen line and whether
// Enter was pressed (run it now) rather than another key (edit it first).
func (e *Editor) search(history []string) (string, bool, error) {
	query := ""
	match := len(history)
	failed := false
	find := func(from int) {
		for i := from; i >= 0; i-- {
			if i < len(history) && query != "" && strings.Contains(history[i], query) {
				match, failed = i, false
				return
			}
		}
		failed = query != ""
	}
	original := string(e.buf)
	for {
		found := original
		if match < len(history) {
			found = history[match]
		}
		label := "(reverse-i-search)"
		if failed {
			label = "(failed reverse-i-search)"
		}
		e.prompt = fmt.Sprintf("%s`%s': ", label, query)
		e.buf, e.pos = []rune(found), strings.Index(found, query)
		if e.pos < 0 || query == "" {
			e.pos = 0
		} else {
			e.pos = len([]rune(found[:e.pos]))
		}
		e.refresh()

		r, _, err := e.in.ReadRune()
		if err != nil {
			return "", false, io.EOF
		}
		switch r {
		case enter, ctrlJ:
			return found, true, nil
		case ctrlC, ctrlG:
			return original, false, nil
		case ctrlR:
			find(match - 1)
		case backspace, ctrlH:
			if query != "" {
				runes := []rune(query)
				query = string(runes[:len(runes)-1])
				match = len(history)
				find(len(history) - 1)
			}
		case escape:
			e.escapeSequence()
			return found, false, nil
		default:
			if unicode.IsPrint(r) {
				query += string(r)
				find(match)
			} else {
				return found, false, nil
			}
		}
	}
}

func (e *Editor) setLine(line string) {
	e.buf = []rune(line)
	e.pos = len(e.buf)
}

func (e *Editor) insert(runes ...rune) {
	tail := append([]rune{}, e.buf[e.pos:]...)
	e.buf = append(append(e.buf[:e.pos], runes...), tail...)
	e.pos += len(runes)
}

func (e *Editor) deleteAt(i int) {
	if i < len(e.buf) {
		e.buf = append(e.buf[:i], e.buf[i+1:]...)
	}
}

func (e *Editor) move(delta int) {
	e.pos = clamp(e.pos+delta, 0, len(e.buf))
}

// clamp limits value to low..high (go.mod targets Go 1.18, which has no min and max builtins).
func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func (e *Editor) wordStart() int {
	i := e.pos
	for i > 0 && !isWordRune(e.buf[i-1]) {
		i--
	}
	for i > 0 && isWordRune(e.buf[i-1]) {
		i--
	}
	return i
}

func (e *Editor) wordEnd() int {
	i := e.pos
	for i < len(e.buf) && !isWordRune(e.buf[i]) {
		i++
	}
	for i < len(e.buf) && isWordRune(e.buf[i]) {
		i++
	}
	return i
}

// refresh redraws the line. A line longer than the terminal scrolls sideways to keep the cursor visible.
func (e *Editor) refresh() {
	promptWidth := len([]rune(e.prompt))
	visible := e.width() - promptWidth - 1
	if visible < 1 {
		visible = 1
	}
	if e.pos < e.offset {
		e.offset = e.pos
	}
	if e.pos > e.offset+visible {
		e.offset = e.pos - visible
	}
	if e.offset > 0 && len(e.buf)-e.offset < visible {
		e.offset = clamp(len(e.buf)-visible, 0, len(e.buf))
	}
	end := clamp(e.offset+visible, 0, len(e.buf))
	var b strings.Builder
	b.WriteString("\r")
	b.WriteString(e.prompt)
	b.WriteString(string(e.buf[e.offset:end]))
	b.WriteString("\x1b[K\r")
	if column := promptWidth + e.pos - e.offset; column > 0 {
		fmt.Fprintf(&b, "\x1b[%dC", column)
	}
	io.WriteString(e.out, b.String())
}

// finish shows the whole line (it may have been scrolled) and moves to the next line.
func (e *Editor) finish() {
	e.pos, e.offset = len(e.buf), 0
	io.WriteString(e.out, "\r"+e.prompt+string(e.buf)+"\x1b[K\r\n")
}
