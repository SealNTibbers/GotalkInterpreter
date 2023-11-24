package talkio

import (
	"errors"
	"strings"
)

type StringWriter struct {
	strings.Builder
}

func (w *StringWriter) GetPosition() int64 {
	return (int64)(w.Len())
}

func (w *StringWriter) SetPosition(position int64) error {
	if position > (int64)(w.Len()) {
		return errors.New("strings.StringWriter.SetPosition: invalid position")
	}
	string := w.String()[:position]
	w.Reset()
	w.WriteString(string)
	return nil
}
