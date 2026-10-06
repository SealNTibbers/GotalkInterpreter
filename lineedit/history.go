package lineedit

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// History is the list of entered lines, oldest first, optionally kept in a file (one entry per line).
type History struct {
	mu    sync.Mutex
	lines []string
	path  string
	max   int
}

const defaultHistorySize = 1000

// NewHistory returns an in-memory history.
func NewHistory() *History {
	return &History{max: defaultHistorySize}
}

// LoadHistory reads the history file at path (a missing file is fine) and appends new entries to it.
// It keeps the last max entries (1000 when max <= 0).
func LoadHistory(path string, max int) *History {
	if max <= 0 {
		max = defaultHistorySize
	}
	h := &History{path: path, max: max}
	file, err := os.Open(path)
	if err != nil {
		return h
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	total := 0
	for scanner.Scan() {
		if line := scanner.Text(); strings.TrimSpace(line) != "" {
			h.lines = append(h.lines, line)
			total++
		}
	}
	file.Close()
	if len(h.lines) > max {
		h.lines = append([]string{}, h.lines[len(h.lines)-max:]...)
	}
	if total > 2*max { // compact a file that has grown well past the limit
		h.rewrite()
	}
	return h
}

// Add appends a line, unless it is blank or repeats the last entry. Lines must not contain newlines.
func (h *History) Add(line string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if strings.TrimSpace(line) == "" || strings.ContainsAny(line, "\r\n") {
		return
	}
	if len(h.lines) > 0 && h.lines[len(h.lines)-1] == line {
		return
	}
	h.lines = append(h.lines, line)
	if len(h.lines) > h.max {
		h.lines = h.lines[len(h.lines)-h.max:]
	}
	if h.path != "" {
		if file, err := os.OpenFile(h.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			file.WriteString(line + "\n")
			file.Close()
		}
	}
}

// Entries returns a copy of the history, oldest first.
func (h *History) Entries() []string {
	return h.entries()
}

func (h *History) entries() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string{}, h.lines...)
}

func (h *History) rewrite() {
	temporary := h.path + ".tmp"
	if err := os.WriteFile(temporary, []byte(strings.Join(h.lines, "\n")+"\n"), 0o600); err == nil {
		os.Rename(temporary, h.path)
	}
}
