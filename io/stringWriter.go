package io

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

func (w *StringWriter) SetPosition(position int64) (int64, error) {
	if position > (int64)(w.Len()) {
		return -1, errors.New("strings.StringWriter.SetPosition: invalid position")
	}
	string := w.String()[:position]
	w.Reset()
	w.WriteString(string)
	return position, nil
}
