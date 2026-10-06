//go:build linux

package lineedit

import "syscall"

const (
	getTermios = syscall.TCGETS
	setTermios = syscall.TCSETS
)
