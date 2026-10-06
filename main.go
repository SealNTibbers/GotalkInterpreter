package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/SealNTibbers/GotalkInterpreter/evaluator"
	"github.com/SealNTibbers/GotalkInterpreter/lineedit"
	"github.com/SealNTibbers/GotalkInterpreter/parser"
	"github.com/SealNTibbers/GotalkInterpreter/treeNodes"
)

const usage = `Usage: gotalk [options] [file.st ...]

Without arguments, starts an interactive Smalltalk console (a workspace: variables you assign stay defined).
Files are run in order, in the same workspace. "-" reads a program from standard input.
Programs print with Transcript (show:, cr, print:, display:, tab, space, showCr:).

Options:
`

func main() {
	expression := flag.String("e", "", "evaluate `expression`, print its result and exit (after running the files)")
	interactive := flag.Bool("i", false, "start the interactive console after running the files")
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	console := newConsole(os.Stdout, os.Stderr)
	for _, path := range flag.Args() {
		if path == "-" {
			console.runReader("stdin", os.Stdin)
		} else {
			console.runFile(path)
		}
	}
	if *expression != "" {
		console.printIt(*expression)
	}

	ranSomething := flag.NArg() > 0 || *expression != ""
	switch {
	case *interactive || !ranSomething && isTerminal(os.Stdin):
		console.repl(os.Stdin)
	case !ranSomething:
		// gotalk < program.st
		console.runReader("stdin", os.Stdin)
	}
	if console.failed {
		os.Exit(1)
	}
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

const (
	prompt             = "st> "
	continuationPrompt = "  > "
)

const help = `Type Smalltalk expressions; the result of each input is printed (print it).
Statements are separated by periods, and input continues on the next line while a bracket, string or
comment is open or an expression is unfinished. Variables you assign stay defined.

  Transcript show: 'Hello'; cr.      print text (also print:, display:, showCr:, tab, space)
  :load file.st                      run a file in this workspace
  :reset                             forget all variables
  :history                           list earlier inputs
  :help                              this text
  :quit                              exit (or Ctrl+D); Ctrl+C cancels the current input

Editing: Left/Right, Home/End (Ctrl+A/E), Alt+Left/Right or Alt+B/F move by word; Backspace, Delete,
Ctrl+K/U delete to the end/start, Ctrl+W deletes a word; Up/Down recall earlier inputs, Ctrl+R searches
them. The history is kept in ~/.gotalk_history (set GOTALK_HISTORY to change it).
`

// console runs Smalltalk programs in one workspace, from files or interactively.
type console struct {
	vm      *evaluator.Evaluator
	out     *lineWriter
	errOut  io.Writer
	failed  bool              // a program in a file or -e failed
	history *lineedit.History // the interactive history, nil when not on a terminal
}

func newConsole(out, errOut io.Writer) *console {
	c := &console{out: &lineWriter{w: out, atLineStart: true}, errOut: errOut}
	c.reset()
	return c
}

func (c *console) reset() {
	c.vm = evaluator.NewSmalltalkWorkspace()
	c.vm.SetVar("Transcript", &transcript{out: c.out})
}

// evaluate runs a program; a panic in the interpreter is reported as an error instead of ending the session.
func (c *console) evaluate(code string) (result treeNodes.SmalltalkObjectInterface, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal error: %v", r)
		}
	}()
	return c.vm.Evaluate(code)
}

// printIt evaluates code and prints its result, like "print it" in a Smalltalk workspace.
func (c *console) printIt(code string) bool {
	result, err := c.evaluate(code)
	if err != nil {
		c.reportError("", err)
		return false
	}
	if _, isTranscript := result.(*transcript); isTranscript || result == nil {
		c.out.endLine()
		return true
	}
	c.out.endLine()
	fmt.Fprintln(c.out, printString(result))
	return true
}

func (c *console) reportError(where string, err error) {
	c.out.endLine()
	fmt.Fprintf(c.errOut, "%sError: %v\n", where, err)
}

func printString(object treeNodes.SmalltalkObjectInterface) string {
	printed, err := object.Perform("printString", nil)
	if err != nil {
		return "a " + object.TypeOf()
	}
	return printed.(*treeNodes.SmalltalkString).GetValue()
}

func (c *console) runFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: %v\n", err)
		c.failed = true
		return
	}
	defer file.Close()
	c.runReader(path, file)
}

// runReader runs a program silently: only Transcript output and errors are printed. A file can be split into
// chunks ending with '!' (the classic file-in format); each chunk runs even if an earlier one failed.
func (c *console) runReader(name string, reader io.Reader) {
	source, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(c.errOut, "Error: %s: %v\n", name, err)
		c.failed = true
		return
	}
	text := string(source)
	if strings.HasPrefix(text, "#!") { // #!/usr/bin/env gotalk; keep the newline so line numbers stay right
		if newline := strings.IndexByte(text, '\n'); newline >= 0 {
			text = text[newline:]
		} else {
			text = ""
		}
	}
	for _, chunk := range splitChunks(text) {
		if isBlank(chunk.code) {
			continue
		}
		if _, err := c.evaluate(chunk.code); err != nil {
			c.reportError(fmt.Sprintf("%s:%d: ", name, chunk.line), err)
			c.failed = true
		}
	}
	c.out.endLine()
}

// lineReader is where the console gets input: a line editor on a terminal, plain lines otherwise.
type lineReader interface {
	ReadLine(prompt string) (string, error)
}

// historyFile is where the console keeps its history: $GOTALK_HISTORY, or ~/.gotalk_history.
// GOTALK_HISTORY= (empty) keeps it in memory only.
func historyFile() string {
	if path, ok := os.LookupEnv("GOTALK_HISTORY"); ok {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gotalk_history")
}

// repl is the interactive console. On a terminal, lines are edited with the arrow keys and Up/Down (or Ctrl+R)
// recall earlier inputs.
func (c *console) repl(input io.Reader) {
	fmt.Fprintln(c.out, "GotalkInterpreter Smalltalk. Type :help for help, Ctrl+D to exit.")
	var reader lineReader
	if file, ok := input.(*os.File); ok && isTerminal(file) {
		editor := lineedit.New(file, c.out)
		if path := historyFile(); path != "" {
			editor.History = lineedit.LoadHistory(path, 0)
		}
		c.history = editor.History
		reader = editor
	} else {
		reader = &plainReader{in: bufio.NewReader(input), out: c.out}
	}
	// Ctrl+C can't stop an evaluation; while one runs, it is ignored instead of ending the session.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)

	var pending []string
	for {
		currentPrompt := prompt
		if len(pending) > 0 {
			currentPrompt = continuationPrompt
		}
		line, err := reader.ReadLine(currentPrompt)
		c.out.atLineStart = true // the user's Enter ended the prompt's line
		if err == lineedit.ErrInterrupted {
			pending = nil
			continue
		}
		if err != nil {
			fmt.Fprintln(c.out)
			return
		}
		if len(pending) == 0 && strings.HasPrefix(strings.TrimSpace(line), ":") {
			c.remember([]string{line})
			if quit := c.command(strings.TrimSpace(line)); quit {
				return
			}
			continue
		}
		pending = append(pending, line)
		code := strings.Join(pending, "\n")
		if needsMoreInput(code) {
			continue
		}
		c.remember(pending)
		pending = nil
		code = strings.TrimSuffix(strings.TrimSpace(code), "!") // habit from chunk-format consoles
		if !isBlank(code) {
			c.printIt(code)
		}
		select {
		case <-interrupts:
		default:
		}
	}
}

// remember adds an input to the history. An input typed on several lines becomes one entry, so Up recalls it
// whole, unless a line break is inside a string or comment, where it matters; then each line is an entry.
func (c *console) remember(lines []string) {
	if c.history == nil {
		return
	}
	code := strings.Join(lines, "\n")
	if lineBreaksInCode(code) {
		parts := make([]string, 0, len(lines))
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				parts = append(parts, strings.TrimSpace(line))
			}
		}
		c.history.Add(strings.Join(parts, " "))
		return
	}
	for _, line := range lines {
		c.history.Add(line)
	}
}

// lineBreaksInCode tells whether every line break in code is outside strings and comments.
func lineBreaksInCode(code string) bool {
	scan := newLexState()
	runes := []rune(code)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\n' && !scan.isCode() {
			return false
		}
		scan.feed(runes, &i)
	}
	return true
}

// plainReader reads lines without editing, for piped input.
type plainReader struct {
	in  *bufio.Reader
	out io.Writer
}

func (r *plainReader) ReadLine(prompt string) (string, error) {
	fmt.Fprint(r.out, prompt)
	line, err := r.in.ReadString('\n')
	if line == "" && err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// command runs a console command (:help, :load, ...) and tells whether to quit.
func (c *console) command(line string) bool {
	name, argument, _ := strings.Cut(line, " ")
	argument = strings.TrimSpace(argument)
	switch name {
	case ":quit", ":q", ":exit":
		return true
	case ":help", ":h", ":?":
		fmt.Fprint(c.out, help)
	case ":history":
		if c.history == nil {
			break
		}
		for i, entry := range c.history.Entries() {
			fmt.Fprintf(c.out, "%5d  %s\n", i+1, entry)
		}
	case ":reset":
		c.reset()
		fmt.Fprintln(c.out, "All variables forgotten.")
	case ":load", ":l":
		if argument == "" {
			fmt.Fprintln(c.errOut, "Usage: :load file.st")
			break
		}
		c.runFile(argument)
		c.failed = false // a failed file doesn't make the session's exit status fail
	default:
		fmt.Fprintf(c.errOut, "Unknown command %s; type :help\n", name)
	}
	return false
}

type chunk struct {
	code string
	line int // where the chunk starts, 1-based
}

// splitChunks splits source at '!' outside strings and comments. As in the classic chunk format, "!!" stands
// for one '!' everywhere; a single '!' inside a string or comment is kept too.
func splitChunks(source string) []chunk {
	var chunks []chunk
	var current strings.Builder
	line, start := 1, 1
	finish := func() {
		code := current.String()
		leading := code[:len(code)-len(strings.TrimLeft(code, " \t\r\n"))]
		chunks = append(chunks, chunk{code, start + strings.Count(leading, "\n")})
		current.Reset()
		start = line
	}
	scan := newLexState()
	runes := []rune(source)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '!' && i+1 < len(runes) && runes[i+1] == '!' {
			current.WriteRune('!')
			i++
			continue
		}
		if runes[i] == '!' && scan.isCode() {
			finish()
			continue
		}
		scan.feed(runes, &i)
		segment := string(runes[scan.from : i+1])
		current.WriteString(segment)
		line += strings.Count(segment, "\n")
	}
	finish()
	return chunks
}

// needsMoreInput tells whether interactive input continues on the next line: a bracket, string or comment is
// still open, or the parser reached the end in the middle of an expression (`3 +`, `x :=`, `a at:`).
func needsMoreInput(code string) bool {
	scan := newLexState()
	runes := []rune(code)
	for i := 0; i < len(runes); i++ {
		scan.feed(runes, &i)
	}
	if !scan.isCode() || scan.depth > 0 {
		return true
	}
	if scan.depth < 0 {
		return false // let the parser report the extra bracket
	}
	_, err := parser.InitializeParserFor(code)
	return err != nil && strings.Contains(err.Error(), "end of input")
}

// isBlank tells whether code holds nothing but whitespace and comments.
func isBlank(code string) bool {
	scan := newLexState()
	runes := []rune(code)
	for i := 0; i < len(runes); i++ {
		if scan.isCode() && runes[i] != '"' && !unicode.IsSpace(runes[i]) {
			return false
		}
		scan.feed(runes, &i)
	}
	return true
}

// lexState follows strings, comments, character literals and bracket depth through Smalltalk source.
type lexState struct {
	inString, inComment bool
	depth               int
	from                int // first rune consumed by the last feed
}

func newLexState() *lexState {
	return &lexState{}
}

func (s *lexState) isCode() bool {
	return !s.inString && !s.inComment
}

// feed consumes the rune at *i (and the character after a $), moving *i to the last rune consumed.
func (s *lexState) feed(runes []rune, i *int) {
	s.from = *i
	r := runes[*i]
	switch {
	case s.inString:
		if r == '\'' {
			if *i+1 < len(runes) && runes[*i+1] == '\'' {
				*i++
			} else {
				s.inString = false
			}
		}
	case s.inComment:
		if r == '"' {
			s.inComment = false
		}
	case r == '\'':
		s.inString = true
	case r == '"':
		s.inComment = true
	case r == '$':
		if *i+1 < len(runes) {
			*i++
		}
	case r == '(' || r == '[' || r == '{':
		s.depth++
	case r == ')' || r == ']' || r == '}':
		s.depth--
	}
}

// lineWriter remembers whether the output is at the start of a line, so results and prompts don't get
// appended to text a program printed without a newline.
type lineWriter struct {
	w           io.Writer
	atLineStart bool
}

func (l *lineWriter) Write(p []byte) (int, error) {
	if len(p) > 0 {
		l.atLineStart = p[len(p)-1] == '\n'
	}
	return l.w.Write(p)
}

func (l *lineWriter) endLine() {
	if !l.atLineStart {
		fmt.Fprintln(l)
	}
}

// transcript is the Transcript global: a stream on standard output.
type transcript struct {
	out io.Writer
}

func (t *transcript) TypeOf() string {
	return "Transcript"
}

func (t *transcript) Value() treeNodes.SmalltalkObjectInterface {
	return t
}

func (t *transcript) Perform(selector string, args []treeNodes.SmalltalkObjectInterface) (treeNodes.SmalltalkObjectInterface, error) {
	text := func(message string) (string, error) {
		printed, err := args[0].Perform(message, nil)
		if err != nil {
			return "", err
		}
		return printed.(*treeNodes.SmalltalkString).GetValue(), nil
	}
	write := func(s string) (treeNodes.SmalltalkObjectInterface, error) {
		fmt.Fprint(t.out, s)
		return t, nil
	}
	arity := strings.Count(selector, ":")
	if selector == "<<" {
		arity = 1
	}
	if arity != len(args) {
		return nil, fmt.Errorf("#%s expects %d argument(s), got %d", selector, arity, len(args))
	}
	switch selector {
	case "show:", "display:", "nextPutAll:", "<<":
		s, err := text("displayString")
		if err != nil {
			return nil, err
		}
		return write(s)
	case "showCr:", "displayNl:", "crShow:":
		s, err := text("displayString")
		if err != nil {
			return nil, err
		}
		if selector == "crShow:" {
			return write("\n" + s)
		}
		return write(s + "\n")
	case "print:":
		s, err := text("printString")
		if err != nil {
			return nil, err
		}
		return write(s)
	case "cr", "nl", "lf", "newLine":
		return write("\n")
	case "tab":
		return write("\t")
	case "space":
		return write(" ")
	case "flush", "endEntry", "yourself", "value":
		return t, nil
	case "printString", "displayString":
		return treeNodes.NewSmalltalkString("a Transcript"), nil
	case "isNil":
		return treeNodes.NewSmalltalkBoolean(false), nil
	case "notNil":
		return treeNodes.NewSmalltalkBoolean(true), nil
	}
	return nil, fmt.Errorf("Transcript does not understand #%s", selector)
}
