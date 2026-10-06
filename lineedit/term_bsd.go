//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package lineedit

import "syscall"

const (
	getTermios = syscall.TIOCGETA
	setTermios = syscall.TIOCSETA
)
