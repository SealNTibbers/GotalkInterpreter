package scanner

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/SealNTibbers/GotalkInterpreter/talkio"
)

const (
	EOF       = "#eof"
	ALPHABET  = "#alphabetic"
	DIGIT     = "#digit"
	BIN       = "#binary"
	SPEC      = "#special"
	SEPARATOR = "#separator"

	BOOLEAN = "boolean"
	NIL     = "nil"
	STRING  = "string"
	NUMBER  = "number"
	IDENT   = "identifier"
	SYMBOL  = "symbol"
	CHAR    = "character"
	ARRAY   = "array"
	KEYWORD = "keyword"
)

func New(input talkio.StringReader) *Scanner {

	scanner := &Scanner{}

	scanner.initializeBuffer()

	scanner.on(input)

	scanner.step()

	scanner.stripSeparators()

	return scanner

}

type Scanner struct {
	buffer              *talkio.StringWriter
	stream              *talkio.StringReader
	classificationTable []string
	characterType       string
	currentCharacter    rune
	tokenStart          int64
	token               TokenInterface
	err                 error
}

func (s *Scanner) on(input talkio.StringReader) {
	s.buffer.Grow(60)
	s.stream = &input

	//classification table init
}

func (s *Scanner) step() rune {
	if s.stream.AtEnd() {
		s.characterType = EOF
		s.currentCharacter = 0
		return s.currentCharacter
	}

	s.currentCharacter, _, _ = s.stream.ReadRune()
	s.characterType = s.classify(s.currentCharacter)
	return s.currentCharacter
}

// stripSeparators skips whitespace and "comments". An unterminated comment is reported by the next call to Next.
func (s *Scanner) stripSeparators() {
	for {
		if s.characterType == SEPARATOR && s.characterType != EOF {
			s.step()
		} else if s.currentCharacter == '"' {
			start := s.stream.GetPosition()
			s.step()
			for s.currentCharacter != '"' {
				if s.characterType == EOF {
					s.err = fmt.Errorf("unterminated comment starting at %d", start)
					return
				}
				s.step()
			}
			s.step()
		} else {
			break
		}
	}
}

func (s *Scanner) getClassificationTable() []string {
	if s.classificationTable == nil {
		s.initializeClassificationTable()
	}
	return s.classificationTable
}

func (s *Scanner) initializeBuffer() {
	s.buffer = &talkio.StringWriter{}
}

func (s *Scanner) initializeClassificationTable() []string {
	s.classificationTable = make([]string, 255)
	var i rune
	for i = 0; i < 255; i++ {
		if unicode.IsLetter(i) {
			s.classificationTable[i] = ALPHABET
		}

		if unicode.IsSpace(i) {
			s.classificationTable[i] = SEPARATOR
		}

		if unicode.IsNumber(i) {
			s.classificationTable[i] = DIGIT
		}
	}
	s.classificationTable['_'] = ALPHABET

	s.initializeRuneTypes(`!%&*+,-/<=>?@\~|`, BIN)

	s.classificationTable[177] = BIN
	s.classificationTable[183] = BIN
	s.classificationTable[215] = BIN
	s.classificationTable[247] = BIN

	s.initializeRuneTypes(`().:;[]^`, SPEC)

	return s.classificationTable
}

func (s *Scanner) initializeRuneTypes(runes string, symbol string) {
	for _, character := range runes {
		s.classificationTable[character] = symbol
	}
}

func (s *Scanner) classify(character rune) string {
	if character == 0 {
		return SEPARATOR
	}
	if character > 255 {
		if unicode.IsLetter(character) {
			return ALPHABET
		} else {
			if unicode.IsSpace(character) {
				return SEPARATOR
			} else {
				return ""
			}
		}
	}
	return s.getClassificationTable()[character]
}

func (s *Scanner) Next() (TokenInterface, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.buffer.Reset()
	s.tokenStart = s.stream.GetPosition()
	if s.characterType == EOF {
		s.token = &EOFToken{&Token{s.tokenStart + 1}}
	} else {
		sT, err := s.scanToken()
		if err != nil {
			return nil, err
		}
		s.token = sT
	}
	s.stripSeparators()
	return s.token, nil
}

func (s *Scanner) previousStepPosition() int64 {
	if s.characterType == EOF {
		return s.stream.GetPosition()
	} else {
		return s.stream.GetPosition() - 1
	}
}

func (s *Scanner) scanToken() (TokenInterface, error) {
	if s.characterType == ALPHABET {
		return s.scanIdentifierOrKeyword(), nil
	}

	if s.characterType == DIGIT || (s.currentCharacter == '-' && s.classify(s.stream.PeekRune()) == DIGIT) {
		return s.scanNumber()
	}

	if s.characterType == BIN {
		return s.scanBinaryInSelector(), nil
	}

	if s.characterType == SPEC {
		return s.scanSpecialCharacter(), nil
	}

	if s.currentCharacter == '\'' {
		return s.scanStringSymbol()
	}

	if s.currentCharacter == '#' {
		return s.scanLiteral()
	}

	if s.currentCharacter == '$' {
		return s.scanCharacter()
	}

	return nil, fmt.Errorf("unexpected character %q at %d", s.currentCharacter, s.tokenStart)
}

// scanCharacter reads a character literal such as $a. Characters evaluate to one-character strings, as in Amber.
func (s *Scanner) scanCharacter() (TokenInterface, error) {
	if s.stream.AtEnd() {
		return nil, fmt.Errorf("character expected after $ at %d", s.tokenStart)
	}
	s.step()
	value := string(s.currentCharacter)
	s.step()
	return NewLiteralToken(s.tokenStart, s.previousStepPosition(), value, CHAR), nil
}

func (s *Scanner) scanIdentifierOrKeyword() TokenInterface {
	s.scanName()

	if s.currentCharacter == ':' && s.stream.PeekRune() != '=' {
		return s.scanKeyword()
	}
	name := s.buffer.String()
	if name == "true" {
		return NewLiteralToken(s.tokenStart, s.previousStepPosition(), "true", BOOLEAN)
	}
	if name == "false" {
		return NewLiteralToken(s.tokenStart, s.previousStepPosition(), "false", BOOLEAN)
	}
	if name == "nil" {
		return NewLiteralToken(s.tokenStart, s.previousStepPosition(), "nil", NIL)
	}
	return &IdentifierToken{&ValueToken{&Token{s.tokenStart}, name, IDENT}}
}

func (s *Scanner) scanKeyword() TokenInterface {
	var outputPosition, inputPosition int64
	for {
		if s.currentCharacter == ':' {
			s.buffer.WriteRune(s.currentCharacter)
			outputPosition = s.buffer.GetPosition()
			inputPosition = s.stream.GetPosition()
			s.step()

			for {
				if s.characterType == ALPHABET {
					s.scanName()
				} else {
					break
				}
			}
		} else {
			break
		}
	}

	_ = s.buffer.SetPosition(outputPosition)
	_ = s.stream.SetPosition(inputPosition)
	s.step()
	name := s.buffer.String()
	if (strings.Count(name, ":")) == 1 {
		return &KeywordToken{&ValueToken{&Token{s.tokenStart}, name, KEYWORD}}
	} else {
		return &MultiKeywordLiteralToken{NewLiteralToken(s.tokenStart, s.tokenStart+(int64)(len(name)), "#"+name, KEYWORD)}
	}
}

func (s *Scanner) scanName() {
	for {
		if s.characterType == ALPHABET || s.characterType == DIGIT {
			s.buffer.WriteRune(s.currentCharacter)
			s.step()
		} else {
			break
		}
	}
}

func (s *Scanner) scanNumber() (*NumberLiteralToken, error) {
	start := s.stream.GetPosition()

	number, err := s.scanNumberVisualWorks()
	if err != nil {
		return nil, err
	}
	currentPosition := s.stream.GetPosition()

	var stop int64
	if s.characterType == EOF {
		stop = currentPosition
	} else {
		stop = currentPosition - 1
	}
	err = s.stream.SetPosition(start - 1)
	if err != nil {
		return nil, err
	}
	_, err = s.stream.ReadRunes(stop - start + 1)
	if err != nil {
		return nil, errors.New("can't read an amount of runes to scan number")
	}
	err = s.stream.SetPosition(currentPosition)
	if err != nil {
		return nil, err
	}

	return &NumberLiteralToken{NewLiteralToken(start, stop, string(number), NUMBER)}, nil
}

func (s *Scanner) scanNumberVisualWorks() (string, error) {
	err := s.stream.Skip(-1)
	if err != nil {
		return "", err
	}
	number, err := s.readSmalltalkSyntaxFromStream()
	if err != nil {
		return "", err
	}
	s.step()
	return number, nil
}

// readSmalltalkSyntaxFromStream reads a number literal: an optional minus, digits, then either a radix part
// (16r1F) or an optional fraction and exponent (1.5e-3). Decimal values are converted by strconv, so they are
// correctly rounded.
func (s *Scanner) readSmalltalkSyntaxFromStream() (string, error) {
	start := s.stream.GetPosition()
	sign := ""
	if s.stream.PeekRuneFor('-') {
		sign = "-"
	}
	digits := s.readDigits(10)
	if digits == "" {
		return "", fmt.Errorf("digit expected at %d", start)
	}
	if s.stream.PeekRune() == 'r' {
		radix, err := strconv.Atoi(digits)
		if err != nil || radix < 2 || radix > 36 {
			return "", fmt.Errorf("invalid radix %s at %d", digits, start)
		}
		_, _, _ = s.stream.ReadRune()
		radixDigits := s.readDigits(radix)
		if radixDigits == "" {
			return "", fmt.Errorf("digits expected after radix at %d", start)
		}
		value, err := strconv.ParseInt(sign+radixDigits, radix, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number at %d: %v", start, err)
		}
		return strconv.FormatFloat(float64(value), 'f', -1, 64), nil
	}
	mantissa := sign + digits
	if s.stream.PeekRune() == '.' {
		position := s.stream.GetPosition()
		_, _, _ = s.stream.ReadRune()
		if fraction := s.readDigits(10); fraction != "" {
			mantissa += "." + fraction
		} else {
			// "3." ends a statement
			_ = s.stream.SetPosition(position)
		}
	}
	if next := s.stream.PeekRune(); next == 'e' || next == 'd' || next == 'q' {
		position := s.stream.GetPosition()
		_, _, _ = s.stream.ReadRune()
		expSign := ""
		if s.stream.PeekRuneFor('-') {
			expSign = "-"
		}
		if exponent := s.readDigits(10); exponent != "" {
			mantissa += "e" + expSign + exponent
		} else if next == 'e' || expSign != "" {
			// not an exponent but a unary message, e.g. 2e
			_ = s.stream.SetPosition(position)
		}
		// otherwise a bare VisualWorks precision suffix (1.02d), which is consumed
	}
	value, err := strconv.ParseFloat(mantissa, 64)
	if err != nil {
		return "", fmt.Errorf("invalid number %s at %d", mantissa, start)
	}
	return strconv.FormatFloat(value, 'f', -1, 64), nil
}

// readDigits consumes the longest run of digits valid in radix (0-9, then A-Z).
func (s *Scanner) readDigits(radix int) string {
	var digits strings.Builder
	for !s.stream.AtEnd() {
		character := s.stream.PeekRune()
		digit := CharToNum(character)
		if digit < 0 && 'A' <= character && character <= 'Z' {
			digit = int(character-'A') + 10
		}
		if digit < 0 || digit >= radix {
			break
		}
		_, _, _ = s.stream.ReadRune()
		digits.WriteRune(character)
	}
	return digits.String()
}

func CharToNum(r rune) int {
	if '0' <= r && r <= '9' {
		return int(r) - '0'
	}
	return -1
}

func (s *Scanner) scanSpecialCharacter() TokenInterface {
	start := s.stream.GetPosition()
	if s.currentCharacter == ':' {
		s.step()
		if s.currentCharacter == '=' {
			s.step()
			return &AssignmentToken{&Token{start}}
		} else {
			return &SpecialCharacterToken{&ValueToken{&Token{start}, string(':'), SPEC}}
		}
	}
	character := s.currentCharacter
	s.step()
	return &SpecialCharacterToken{&ValueToken{&Token{start}, string(character), SPEC}}
}

func (s *Scanner) scanBinaryInSelector() *BinarySelectorToken {
	s.buffer.WriteRune(s.currentCharacter)
	s.step()
	if s.characterType == BIN && s.currentCharacter != '-' {
		s.buffer.WriteRune(s.currentCharacter)
		s.step()
	}
	val := s.buffer.String()
	binarySelector := &BinarySelectorToken{&ValueToken{&Token{s.tokenStart}, val, BIN}}
	return binarySelector
}

func (s *Scanner) scanBinaryInLiteral() *LiteralToken {
	s.buffer.WriteRune(s.currentCharacter)
	s.step()
	if s.characterType == BIN && s.currentCharacter != '-' {
		s.buffer.WriteRune(s.currentCharacter)
		s.step()
	}
	val := s.buffer.String()
	binarySelector := &LiteralToken{&ValueToken{&Token{s.tokenStart}, val, STRING}, s.previousStepPosition()}
	return binarySelector
}

func (s *Scanner) scanLiteralString() (*LiteralToken, error) {
	s.step()

	for !(s.currentCharacter == '\'' && s.step() != '\'') {
		if s.characterType == EOF {
			return nil, errors.New("UnmatchedQuoteInString")
		}
		s.buffer.WriteRune(s.currentCharacter)
		s.step()
	}

	return &LiteralToken{&ValueToken{&Token{s.tokenStart}, s.buffer.String(), STRING}, s.previousStepPosition()}, nil

}

func (s *Scanner) scanStringSymbol() (*LiteralToken, error) {
	return s.scanLiteralString()
}

func (s *Scanner) scanLiteral() (TokenInterface, error) {
	s.step()
	if s.characterType == BIN {
		binary := s.scanBinaryInLiteral()
		return binary, nil
	}
	if s.currentCharacter == '\'' {
		return s.scanLiteralString()
	}
	if s.currentCharacter == '(' || s.currentCharacter == '[' {
		return s.scanLiteralArrayToken(), nil
	}
	if s.characterType == ALPHABET {
		return s.scanSymbol(), nil
	}
	return nil, fmt.Errorf("literal expected after # at %d", s.tokenStart)
}

// scanSymbol reads #name, #name: or #name:with:. Symbols evaluate to strings, as in Amber.
func (s *Scanner) scanSymbol() *LiteralToken {
	for s.characterType == ALPHABET || s.characterType == DIGIT || s.currentCharacter == ':' {
		s.buffer.WriteRune(s.currentCharacter)
		s.step()
	}
	return NewLiteralToken(s.tokenStart, s.previousStepPosition(), s.buffer.String(), SYMBOL)
}

func (s *Scanner) scanLiteralArrayToken() *LiteralArrayToken {
	valueString := string('#') + string(s.currentCharacter)
	token := &LiteralArrayToken{&ValueToken{&Token{s.tokenStart}, valueString, ARRAY}}
	s.step()
	return token
}

func NewLiteralToken(start int64, stop int64, value string, valueType string) *LiteralToken {
	return &LiteralToken{&ValueToken{&Token{start}, value, valueType}, stop}
}

func NewBinarySelectorToken(start int64, value string) *BinarySelectorToken {
	return &BinarySelectorToken{&ValueToken{&Token{start}, value, BIN}}
}

type TokenInterface interface {
	length() int64
	TypeOfToken() string

	GetStart() int64
	SetStart(int64)
	GetStop() int64
	IsBinary() bool
	IsIdentifier() bool
	IsSpecial() bool
	IsAssignment() bool
	IsLiteralToken() bool
	IsLiteralArrayToken() bool
	IsKeyword() bool
	IsForByteArray() bool
}

type Token struct {
	sourcePointer int64
}

func (t *Token) length() int64 {
	return 0
}

func (t *Token) TypeOfToken() string {
	return "Token"
}

func (t *Token) SetStart(start int64) {
	t.sourcePointer = start
}

func (t *Token) GetStart() int64 {
	return t.sourcePointer
}

func (t *Token) GetStop() int64 {
	return t.GetStart() + t.length() - 1
}

func (t *Token) IsBinary() bool {
	return false
}

func (t *Token) IsIdentifier() bool {
	return false
}

func (t *Token) IsLiteralToken() bool {
	return false
}

func (t *Token) IsKeyword() bool {
	return false
}

func (t *Token) IsSpecial() bool {
	return false
}

func (t *Token) IsLiteralArrayToken() bool {
	return false
}

func (t *Token) IsForByteArray() bool {
	return false
}

func (t *Token) IsAssignment() bool {
	return false
}

type EOFToken struct {
	*Token
}

func (t *EOFToken) TypeOfToken() string {
	return "EOFToken"
}

type ValueTokenInterface interface {
	TokenInterface
	ValueOfToken() string
}

type ValueToken struct {
	*Token
	value     string
	valueType string
}

func (t *ValueToken) length() int64 {
	return (int64)(len(t.value))
}

func (t *ValueToken) TypeOfToken() string {
	return t.valueType
}

func (t *ValueToken) ValueOfToken() string {
	return t.value
}

func (t *ValueToken) SetValue(value string) {
	t.value = value
}

type AssignmentToken struct {
	*Token
}

func (t *AssignmentToken) IsAssignment() bool {
	return true
}

func (t *AssignmentToken) length() int64 {
	return 2
}

type IdentifierToken struct {
	*ValueToken
}

func (i *IdentifierToken) IsIdentifier() bool {
	return true
}

type KeywordToken struct {
	*ValueToken
}

func (k *KeywordToken) IsKeyword() bool {
	return true
}

type LiteralTokenInterface interface {
	ValueTokenInterface
	IsMultiKeyword() bool
}

type LiteralToken struct {
	*ValueToken
	stopPosition int64
}

func (t *LiteralToken) IsMultiKeyword() bool {
	return false
}

func (l *LiteralToken) IsLiteralToken() bool {
	return true
}

type MultiKeywordLiteralToken struct {
	*LiteralToken
}

func (m *MultiKeywordLiteralToken) IsMultiKeyword() bool {
	return true
}

type NumberLiteralToken struct {
	*LiteralToken
}

type BinarySelectorToken struct {
	*ValueToken
}

type SpecialCharacterToken struct {
	*ValueToken
}

func (s *SpecialCharacterToken) IsSpecial() bool {
	return true
}

type LiteralArrayToken struct {
	*ValueToken
}

func (t *LiteralArrayToken) IsLiteralArrayToken() bool {
	return true
}

func (t *LiteralArrayToken) IsForByteArray() bool {
	length := len(t.value)
	return t.value[length-1] == '['
}

func (t *BinarySelectorToken) IsBinary() bool {
	return true
}
