package scanner

import (
	"errors"
	"math"
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

func (s *Scanner) stripSeparators() {
	for {
		if s.characterType == SEPARATOR {
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
	s.classificationTable = make([]string, 255, 255)
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

	return &Token{}, nil
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

	s.buffer.SetPosition(outputPosition)
	s.stream.SetPosition(inputPosition)
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
	s.stream.SetPosition(start - 1)

	_, err = s.stream.ReadRunes(stop - start + 1)
	if err != nil {
		return nil, errors.New("can't read an amount of runes to scan number")
	}
	s.stream.SetPosition(currentPosition)

	return &NumberLiteralToken{NewLiteralToken(start, stop, string(number), NUMBER)}, nil
}

func (s *Scanner) scanNumberVisualWorks() (string, error) {
	s.stream.Skip(-1)
	number, err := s.readSmalltalkSyntaxFromStream()
	if err != nil {
		return "", err
	}
	s.step()
	return number, nil
}

func (s *Scanner) readSmalltalkSyntaxFromStream() (string, error) {
	if s.stream.AtEnd() || unicode.IsLetter(s.stream.PeekRune()) {
		return "0", nil
	}
	neg := s.stream.PeekRuneFor('-')
	value, err := s.readIntegerWithRadix(10)
	if err != nil {
		return "", err
	}
	floatValue, err := s.readSmalltalkFloat(value)
	if err != nil {
		return "", err
	}
	if neg {
		floatValue *= -1
	}
	return strconv.FormatFloat(floatValue, 'f', -1, 64), nil
}

func (s *Scanner) readIntegerWithRadix(radix int) (int, error) {
	value := 0
	for {
		if s.stream.AtEnd() {
			return value, nil
		}

		character, _, err := s.stream.ReadRune()
		if err != nil {
			return 0, errors.New("readIntegerWithRadix doesn't work as expected. FeelsBadMan")
		}
		digit := CharToNum(character)
		if digit < 0 || digit >= radix {
			s.stream.Skip(-1)
			return value, nil
		} else {
			value = value*radix + digit
		}
	}
	return value, nil
}

func (s *Scanner) readSmalltalkFloat(integerPart int) (float64, error) {
	var num, den float64
	var atEnd bool
	var possibleCoercionClass rune
	var exp int
	precision := 0
	num = 0.0
	den = 1.0
	exp = 0

	if s.stream.PeekRuneFor('.') {
		if !(s.stream.AtEnd()) && unicode.IsDigit(s.stream.PeekRune()) {
			for {
				atEnd = s.stream.AtEnd()
				if atEnd {
					break
				}
				digit, _, err := s.stream.ReadRune()
				if err != nil {
					return 0.0, err
				}
				if !(unicode.IsDigit(digit)) {
					break
				} else {
					digitValue := CharToNum(digit)
					num = num*10.0 + float64(digitValue)
					precision += 1
				}
			}
			den = math.Pow10(precision)
			if !atEnd {
				s.stream.Skip(-1)
			}
		} else {
			//looks like it's just integer
			s.stream.Skip(-1)
		}
	}

	eChar, err := s.stream.PeekRuneError()
	if err == nil && (eChar == 'e' || eChar == 'd') {
		if eChar != 0 {
			possibleCoercionClass, _, _ = s.stream.ReadRune()
		}

		if possibleCoercionClass != 0 {
			endOfNumber := s.stream.GetPosition()
			neg := false
			if s.stream.PeekRuneFor('-') {
				neg = true
			}
			digit, err := s.stream.PeekRuneError()
			if err == nil && (digit != 0) && unicode.IsDigit(digit) {
				exp, err = s.readIntegerWithRadix(10)
				if err != nil {
					return 0, err
				}
				if neg {
					exp = -1 * exp
				}
			} else {
				s.stream.SetPosition(endOfNumber)
			}
		}
	}

	value := float64(integerPart) + (num / den)
	if exp == 0 {
		return value, nil
	} else {
		return value * math.Pow(10, float64(exp)), nil
	}
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
	return nil, nil
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
