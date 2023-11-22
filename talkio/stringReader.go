package talkio

import (
	"errors"
	"io"
	"unicode/utf8"
)

// A StringReader implements the io.StringReader, io.ReaderAt, io.Seeker, io.WriterTo,
// io.ByteScanner, and io.RuneScanner interfaces by reading
// from a string.
type StringReader struct {
	s        string
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

// Len returns the number of bytes of the unread portion of the
// string.
func (r *StringReader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

// Size returns the original length of the underlying string.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (r *StringReader) Size() int64 { return int64(len(r.s)) }

func (r *StringReader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

func (r *StringReader) ReadAt(b []byte, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("strings.StringReader.ReadAt: negative offset")
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

func (r *StringReader) ReadByte() (byte, error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	b := r.s[r.i]
	r.i++
	return b, nil
}

func (r *StringReader) UnreadByte() error {
	r.prevRune = -1
	if r.i <= 0 {
		return errors.New("strings.StringReader.UnreadByte: at beginning of string")
	}
	r.i--
	return nil
}

func (r *StringReader) ReadRune() (ch rune, size int, err error) {
	if r.i >= int64(len(r.s)) {
		r.prevRune = -1
		return 0, 0, io.EOF
	}
	r.prevRune = int(r.i)
	if c := r.s[r.i]; c < utf8.RuneSelf {
		r.i++
		return rune(c), 1, nil
	}
	ch, size = utf8.DecodeRuneInString(r.s[r.i:])
	r.i += int64(size)
	return
}

func (r *StringReader) ReadRunes(amount int64) ([]rune, error) {
	var i int64
	var runes []rune

	for i = 0; i < amount; i++ {
		character, _, err := r.ReadRune()
		if err != nil {
			return runes, err
		}
		runes = append(runes, character)
	}
	return runes, nil
}

func (r *StringReader) UnreadRune() error {
	if r.prevRune < 0 {
		return errors.New("strings.StringReader.UnreadRune: previous operation was not ReadRune")
	}
	r.i = int64(r.prevRune)
	r.prevRune = -1
	return nil
}

func (r *StringReader) PeekRune() rune {
	character, err := r.PeekRuneError()
	if err != nil {
		panic("error rune peeking")
	}
	return character
}

func (r *StringReader) PeekRuneError() (ch rune, err error) {
	if r.i >= int64(len(r.s)) {
		r.prevRune = -1
		return 0, io.EOF
	}
	if c := r.s[r.i]; c < utf8.RuneSelf {
		return rune(c), nil
	}
	ch, _ = utf8.DecodeRuneInString(r.s[r.i:])
	return
}

func (r *StringReader) PeekRuneFor(character rune) bool {
	if r.AtEnd() {
		return false
	}
	nextRune, _, _ := r.ReadRune()
	if character == nextRune {
		return true
	}
	r.Skip(-1)
	return false
}

func (r *StringReader) GetPosition() int64 {
	return r.i
}

func (r *StringReader) SetPosition(position int64) (int64, error) {
	if position > int64(len(r.s)) || position < 0 {
		return -1, errors.New("strings.StringReader.SetPosition: invalid position")
	}
	r.i = position
	r.prevRune = (int)(r.i - 1)
	return r.i, nil
}

func (r *StringReader) Skip(posOffset int64) (int64, error) {
	position, err := r.SetPosition(r.i + posOffset)
	return position, err
}

func (r *StringReader) AtEnd() bool {
	if r.i >= int64(len(r.s)) {
		return true
	} else {
		return false
	}
}

// WriteTo implements the io.WriterTo interface.
func (r *StringReader) WriteTo(w io.Writer) (n int64, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, nil
	}
	s := r.s[r.i:]
	m, err := io.WriteString(w, s)
	if m > len(s) {
		return 0, errors.New("strings.StringReader.WriteTo: invalid WriteString count")
	}
	r.i += int64(m)
	n = int64(m)
	if m != len(s) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

// Reset resets the StringReader to be reading from s.
func (r *StringReader) Reset(s string) { *r = StringReader{s, 0, -1} }

// NewReader returns a new StringReader reading from s.
// It is similar to bytes.NewBufferString but more efficient and read-only.
func NewReader(s string) *StringReader { return &StringReader{s, 0, -1} }
