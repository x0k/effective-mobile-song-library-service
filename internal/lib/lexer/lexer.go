package lexer

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

var (
	ErrInvalidNumber = errors.New("invalid number")
	ErrInvalidString = errors.New("invalid string")
)

type TokenType int

const (
	Number TokenType = iota
	String
	Separator
	Operator
	Symbol
)

type Token interface {
	Position() int
}

type token struct {
	Pos int
}

func newToken(l *Lexer) token {
	return token{
		Pos: l.pos,
	}
}

func (t token) Position() int {
	return t.Pos
}

type NumberToken struct {
	token
	Value int64
}

func (n NumberToken) Type() TokenType {
	return Number
}

type StringToken struct {
	token
	Value string
}

func (s StringToken) Type() TokenType {
	return String
}

type SeparatorToken struct {
	token
	Value rune
}

func (s SeparatorToken) Type() TokenType {
	return Separator
}

type OperatorToken struct {
	token
	Value int
}

func (o OperatorToken) Type() TokenType {
	return Operator
}

type SymbolToken struct {
	token
	Value string
}

func (s SymbolToken) Type() TokenType {
	return Symbol
}

type state int

const (
	idle state = iota
	numToken
	strToken
	separatorToken
	operatorToken
	symbolToken
)

type Lexer struct {
	str        []rune
	strLen     int
	cursor     int
	separators map[rune]struct{}
	operators  [][]rune
	strQuote   rune

	done        bool
	err         error
	token       Token
	state       state
	pos         int
	buff        []rune
	escaped     bool
	indexes     []int
	nextIndexes []int
}

func New(
	operators []string,
	separators []rune,
	str string,
) *Lexer {
	il := len(operators)
	rInstructions := make([][]rune, il)
	for i, k := range operators {
		rInstructions[i] = []rune(k)
	}
	sMap := make(map[rune]struct{}, len(separators))
	for _, sep := range separators {
		sMap[sep] = struct{}{}
	}
	sl := len(str)
	return &Lexer{
		str:         []rune(str),
		strLen:      sl,
		done:        sl == 0,
		operators:   rInstructions,
		separators:  sMap,
		strQuote:    '"',
		indexes:     make([]int, 0, il),
		nextIndexes: make([]int, 0, il),
	}
}

func (l *Lexer) Next() bool {
	if l.done {
		return false
	}
	l.idle()
	for l.process() {
		if l.done = l.err != nil; l.done {
			return false
		}
		l.cursor++
		if l.done = l.cursor >= l.strLen; l.done {
			break
		}
	}
	if l.done {
		switch l.state {
		case idle:
			return false
		case strToken:
			if l.str[l.strLen-1] != l.strQuote {
				l.err = fmt.Errorf("%w: unclosed string, position %d", ErrInvalidString, l.pos)
				return false
			}
			return true
		case separatorToken:
			return true
		case numToken:
			l.err = l.setNumberToken()
			return true
		case operatorToken:
			l.setOperatorToken()
			return true
		case symbolToken:
			l.setSymbolToken()
			return true
		default:
			panic(fmt.Sprintf("unreachable: unexpected state %d, position %d", l.state, l.cursor))
		}
	}
	return !l.done
}

func (l *Lexer) Token() Token {
	return l.token
}

func (l *Lexer) Err() error {
	return l.err
}

func (l *Lexer) idle() {
	l.state = idle
	l.buff = l.buff[:0]
	l.indexes = l.indexes[:0]
	l.nextIndexes = l.nextIndexes[:0]
}

func (l *Lexer) advance() {
	l.cursor++
	l.done = l.cursor == l.strLen
}

func (l *Lexer) process() bool {
	c := l.str[l.cursor]
	switch l.state {
	case idle:
		if l.isSeparator(c) {
			l.setSeparatorToken(c)
			l.advance()
			return false
		}
		if c == l.strQuote {
			l.err = l.startStr()
		} else if unicode.IsDigit(c) {
			l.err = l.startNum(c)
		} else if l.isOperator(c) {
			l.err = l.startOperator()
		} else if !unicode.IsSpace(c) {
			l.err = l.startSymbol(c)
		}
		return true
	case strToken:
		if c != l.strQuote || l.escaped {
			l.err = l.continueStr(c)
			return true
		}
		l.token = StringToken{
			token: newToken(l),
			Value: string(l.buff),
		}
		l.advance()
		return false
	case numToken:
		if unicode.IsDigit(c) {
			l.err = l.continueNum(c)
			return true
		}
		if unicode.IsSpace(c) {
			l.err = l.setNumberToken()
			l.advance()
			return false
		}
		if l.isSeparator(c) {
			l.err = l.setNumberToken()
			return false
		}
		l.err = l.numToSymbol(c)
		return true
	case operatorToken:
		if l.isOperatorContinuation(c) {
			l.err = l.continueOperator()
			return true
		}
		if unicode.IsSpace(c) {
			l.setOperatorToken()
			l.advance()
			return false
		}
		if l.isSeparator(c) {
			l.setOperatorToken()
			return false
		}
		l.err = l.operatorToSymbol(c)
		return true
	case symbolToken:
		if l.isSeparator(c) {
			l.setSymbolToken()
			return false
		}
		if unicode.IsSpace(c) {
			l.setSymbolToken()
			l.advance()
			return false
		}
		l.err = l.continueSymbol(c)
		return true
	default:
		panic(fmt.Sprintf("unreachable: emit on invalid state %d, at position %d", l.state, l.cursor))
	}
}

func (l *Lexer) isSeparator(c rune) bool {
	_, ok := l.separators[c]
	return ok
}

func (l *Lexer) setSeparatorToken(c rune) {
	l.state = separatorToken
	l.pos = l.cursor
	l.token = SeparatorToken{
		token: newToken(l),
		Value: c,
	}
}

func (l *Lexer) startStr() error {
	l.state = strToken
	l.pos = l.cursor
	return nil
}

func (l *Lexer) continueStr(v rune) error {
	l.escaped = v == '\\' && !l.escaped
	if !l.escaped {
		l.buff = append(l.buff, v)
	}
	return nil
}

func (l *Lexer) startNum(c rune) error {
	if c == '0' {
		return fmt.Errorf("%w: leading zero, position %d", ErrInvalidNumber, l.cursor)
	}
	l.state = numToken
	l.pos = l.cursor
	l.buff = append(l.buff, c)
	return nil
}

func (l *Lexer) continueNum(c rune) error {
	l.buff = append(l.buff, c)
	return nil
}

func (l *Lexer) numToSymbol(c rune) error {
	l.state = symbolToken
	l.buff = append(l.buff, c)
	return nil
}

func (l *Lexer) setNumberToken() error {
	n, err := strconv.ParseInt(string(l.buff), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: failed to parse number %v, position %d", ErrInvalidNumber, err, l.pos)
	}
	l.token = NumberToken{
		token: newToken(l),
		Value: n,
	}
	return nil
}

func (l *Lexer) isOperator(c rune) bool {
	for i, instr := range l.operators {
		if instr[0] == c {
			l.indexes = append(l.indexes, i)
		}
	}
	return len(l.indexes) > 0
}

func (l *Lexer) startOperator() error {
	l.state = operatorToken
	l.pos = l.cursor
	l.nextIndexes = l.nextIndexes[:len(l.indexes)]
	return nil
}

func (l *Lexer) isOperatorContinuation(c rune) bool {
	shift := 0
	charIdx := l.cursor - l.pos
	for i, idx := range l.indexes {
		if charIdx >= len(l.operators[idx]) || l.operators[idx][charIdx] != c {
			shift++
		} else {
			l.nextIndexes[i-shift] = l.indexes[i]
		}
	}
	l.nextIndexes = l.nextIndexes[:len(l.indexes)-shift]
	return len(l.nextIndexes) > 0
}

func (l *Lexer) continueOperator() error {
	copy(l.indexes, l.nextIndexes)
	l.indexes = l.indexes[:len(l.nextIndexes)]
	return nil
}

func (l *Lexer) operatorToBuffer() {
	l.state = symbolToken
	ln := l.cursor - l.pos
	if cap(l.buff) < ln {
		l.buff = make([]rune, ln)
	} else {
		l.buff = l.buff[:ln]
	}
	copy(l.buff, l.operators[l.indexes[0]][:ln])
}

func (l *Lexer) operatorToSymbol(c rune) error {
	l.operatorToBuffer()
	l.buff = append(l.buff, c)
	return nil
}

func (l *Lexer) findOperatorIndex() int {
	ln := l.cursor - l.pos
	for _, idx := range l.indexes {
		if len(l.operators[idx]) == ln {
			return idx
		}
	}
	return -1
}

func (l *Lexer) setOperatorToken() {
	idx := l.findOperatorIndex()
	if idx == -1 {
		l.operatorToBuffer()
		l.setSymbolToken()
		return
	}
	l.token = OperatorToken{
		token: newToken(l),
		Value: idx,
	}
}

func (l *Lexer) startSymbol(c rune) error {
	l.state = symbolToken
	l.pos = l.cursor
	l.buff = append(l.buff, c)
	return nil
}

func (l *Lexer) continueSymbol(c rune) error {
	l.buff = append(l.buff, c)
	return nil
}

func (l *Lexer) setSymbolToken() {
	l.token = SymbolToken{
		token: newToken(l),
		Value: string(l.buff),
	}
}
