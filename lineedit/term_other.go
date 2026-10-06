//go:build !(darwin || freebsd || netbsd || openbsd || dragonfly || linux)

package lineedit

import "errors"

// Line editing needs a Unix terminal; elsewhere the editor reads plain lines.

func isTerminal(fd int) bool {
	return false
}

func enableRawMode(fd int) (func(), error) {
	return nil, errors.New("raw mode is not supported")
}

func terminalWidth(fd int) int {
	return 80
}
