//go:build darwin || freebsd || netbsd || openbsd || dragonfly || linux

package lineedit

import (
	"syscall"
	"unsafe"
)

func ioctl(fd int, request uintptr, arg unsafe.Pointer) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), request, uintptr(arg)); errno != 0 {
		return errno
	}
	return nil
}

func isTerminal(fd int) bool {
	var termios syscall.Termios
	return ioctl(fd, getTermios, unsafe.Pointer(&termios)) == nil
}

// enableRawMode turns off line buffering, echo and signal keys, so the editor sees every key. Output
// processing stays on, so "\n" still starts a new line.
func enableRawMode(fd int) (restore func(), err error) {
	var original syscall.Termios
	if err := ioctl(fd, getTermios, unsafe.Pointer(&original)); err != nil {
		return nil, err
	}
	raw := original
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := ioctl(fd, setTermios, unsafe.Pointer(&raw)); err != nil {
		return nil, err
	}
	return func() { ioctl(fd, setTermios, unsafe.Pointer(&original)) }, nil
}

func terminalWidth(fd int) int {
	var size struct{ rows, columns, x, y uint16 }
	if err := ioctl(fd, syscall.TIOCGWINSZ, unsafe.Pointer(&size)); err != nil || size.columns == 0 {
		return 80
	}
	return int(size.columns)
}
