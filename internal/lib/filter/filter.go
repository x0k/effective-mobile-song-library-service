package filter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/lexer"
)

const (
	openParenSep  = '('
	closeParenSep = ')'
	commaSep      = ','
)

var separators = []rune{
	openParenSep,
	closeParenSep,
	commaSep,
}

const (
	equalOp          = "EQ"
	inOp             = "IN"
	greaterOp        = "GT"
	lessOp           = "LT"
	greaterOrEqualOp = "GTE"
	lessOrEqualOp    = "LTE"
	andOp            = "AND"
	orOp             = "OR"
	notOp            = "NOT"
	likeOp           = "LIKE"
	aLikeOp          = "ALIKE"
	dateOp           = "DATE"
)

var operators = []string{
	equalOp,
	inOp,
	greaterOp,
	lessOp,
	greaterOrEqualOp,
	lessOrEqualOp,
	andOp,
	orOp,
	notOp,
	likeOp,
	aLikeOp,
	dateOp,
}

type Filter struct {
	table       string
	schema      map[string]ValueType
	dateFactory func(string) (any, error)
}

func New(
	table string,
	schema map[string]ValueType,
	dateFactory func(string) (any, error),
) *Filter {
	return &Filter{
		table:       table,
		schema:      schema,
		dateFactory: dateFactory,
	}
}

func (p *Filter) Parse(str string) (Expr, error) {
	l := lexer.New(operators, separators, str)
	expr, err := p.parse(l)
	if err != nil {
		return nil, err
	}
	if expr.Type() != BoolType {
		return nil, fmt.Errorf("%w: expected predicate expression", ErrInvalidExpression)
	}
	return expr, nil
}

var ErrInvalidExpression = errors.New("invalid expression")

type ValueType string

const (
	NumberType ValueType = "NUMBER"
	StringType ValueType = "STRING"
	DateType   ValueType = "DATE"
	ArrayType  ValueType = "ARRAY"
	BoolType   ValueType = "BOOL"
)

func ArrayOf(vt ValueType) ValueType {
	return ValueType(fmt.Sprintf("ARRAY(%s)", vt))
}

func isArrayType(vt ValueType) bool {
	return strings.HasPrefix(string(vt), "ARRAY(")
}

func arrayItemType(vt ValueType) ValueType {
	return vt[6 : len(vt)-1]
}

type Expr interface {
	Type() ValueType
	ToSQL(w *strings.Builder, args []any) []any
}

type node struct {
	p   *Filter
	pos int
}

type Number struct {
	node
	val int64
}

func (n Number) Type() ValueType {
	return NumberType
}

func (n Number) ToSQL(w *strings.Builder, args []any) []any {
	args = append(args, n.val)
	w.WriteByte('$')
	w.WriteString(strconv.Itoa(len(args)))
	return args
}

type String struct {
	node
	val string
}

func (s String) Type() ValueType {
	return StringType
}

func (s String) ToSQL(w *strings.Builder, args []any) []any {
	args = append(args, s.val)
	w.WriteByte('$')
	w.WriteString(strconv.Itoa(len(args)))
	return args
}

type Date struct {
	node
	val any
}

func (d Date) Type() ValueType {
	return DateType
}

func (d Date) ToSQL(w *strings.Builder, args []any) []any {
	args = append(args, d.val)
	w.WriteByte('$')
	w.WriteString(strconv.Itoa(len(args)))
	return args
}

type Array struct {
	node
	t    ValueType
	vals []Expr
}

func (a Array) Type() ValueType {
	return a.t
}

func (a Array) ToSQL(w *strings.Builder, args []any) []any {
	w.WriteByte('(')
	for i, v := range a.vals {
		if i > 0 {
			w.WriteString(", ")
		}
		args = v.ToSQL(w, args)
	}
	w.WriteByte(')')
	return args
}

type Column struct {
	node
	t    ValueType
	name string
}

func (c Column) Type() ValueType {
	return c.t
}

func (c Column) ToSQL(w *strings.Builder, args []any) []any {
	w.WriteByte('"')
	w.WriteString(c.p.table)
	w.WriteString("\".\"")
	w.WriteString(c.name)
	w.WriteByte('"')
	return args
}

type unaryOp struct {
	node
	arg Expr
}

type Not unaryOp

func (e Not) Type() ValueType {
	return BoolType
}

func (e Not) ToSQL(w *strings.Builder, args []any) []any {
	w.WriteString("NOT (")
	args = e.arg.ToSQL(w, args)
	w.WriteString(")")
	return args
}

type binaryOp struct {
	node
	left  Expr
	right Expr
}

type Equal binaryOp

func (e Equal) Type() ValueType {
	return BoolType
}

func (e Equal) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	w.WriteString(" = ")
	args = e.right.ToSQL(w, args)
	return args
}

type in binaryOp

func (e in) Type() ValueType {
	return BoolType
}

func (e in) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	_, isCol := e.right.(Column)
	if isCol {
		w.WriteString(" = ANY(")
	} else {
		w.WriteString(" IN ")
	}
	args = e.right.ToSQL(w, args)
	if isCol {
		w.WriteString(")")
	}
	return args
}

type Greater binaryOp

func (e Greater) Type() ValueType {
	return BoolType
}

func (e Greater) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	w.WriteString(" > ")
	args = e.right.ToSQL(w, args)
	return args
}

type Less binaryOp

func (e Less) Type() ValueType {
	return BoolType
}

func (e Less) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	w.WriteString(" < ")
	args = e.right.ToSQL(w, args)
	return args
}

type GreaterOrEqual binaryOp

func (e GreaterOrEqual) Type() ValueType {
	return BoolType
}

func (e GreaterOrEqual) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	w.WriteString(" >= ")
	args = e.right.ToSQL(w, args)
	return args
}

type LessOrEqual binaryOp

func (e LessOrEqual) Type() ValueType {
	return BoolType
}

func (e LessOrEqual) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	w.WriteString(" <= ")
	args = e.right.ToSQL(w, args)
	return args
}

type Like binaryOp

func (e Like) Type() ValueType {
	return BoolType
}

func (e Like) ToSQL(w *strings.Builder, args []any) []any {
	args = e.left.ToSQL(w, args)
	w.WriteString(" ILIKE ")
	args = e.right.ToSQL(w, args)
	return args
}

type ALike binaryOp

func (e ALike) Type() ValueType {
	return BoolType
}

func (e ALike) ToSQL(w *strings.Builder, args []any) []any {
	w.WriteString("EXISTS (SELECT 1 FROM unnest(")
	args = e.left.ToSQL(w, args)
	w.WriteString(") AS element WHERE element ILIKE ")
	args = e.right.ToSQL(w, args)
	w.WriteString(")")
	return args
}

type variadicOp struct {
	node
	args []Expr
}

type And variadicOp

func (e And) Type() ValueType {
	return BoolType
}

func (e And) ToSQL(w *strings.Builder, args []any) []any {
	w.WriteByte('(')
	for i, arg := range e.args {
		if i > 0 {
			w.WriteString(" AND ")
		}
		args = arg.ToSQL(w, args)
	}
	w.WriteByte(')')
	return args
}

type Or variadicOp

func (e Or) Type() ValueType {
	return BoolType
}

func (e Or) ToSQL(w *strings.Builder, args []any) []any {
	w.WriteByte('(')
	for i, arg := range e.args {
		if i > 0 {
			w.WriteString(" OR ")
		}
		args = arg.ToSQL(w, args)
	}
	w.WriteByte(')')
	return args
}

func (p *Filter) node(t lexer.Token) node {
	return node{
		p:   p,
		pos: t.Position(),
	}
}

func (p *Filter) parse(l *lexer.Lexer) (Expr, error) {
	if !l.Next() {
		return nil, fmt.Errorf("%w: unexpected end of expression", ErrInvalidExpression)
	}
	switch t := l.Token().(type) {
	case lexer.NumberToken:
		return Number{
			node: p.node(t),
			val:  t.Value,
		}, nil
	case lexer.StringToken:
		return String{
			node: p.node(t),
			val:  t.Value,
		}, nil
	case lexer.SymbolToken:
		fieldType, ok := p.schema[t.Value]
		if !ok {
			return nil, fmt.Errorf("%w: unknown symbol token %v", ErrInvalidExpression, t)
		}
		return Column{
			node: p.node(t),
			t:    fieldType,
			name: t.Value,
		}, nil
	case lexer.SeparatorToken:
		switch t.Value {
		case commaSep, closeParenSep:
			return nil, fmt.Errorf("%w: unexpected token %v", ErrInvalidExpression, t)
		case openParenSep:
			expressions, err := p.parseList(l)
			if err != nil {
				return nil, err
			}
			if len(expressions) == 0 {
				return nil, fmt.Errorf("%w: unexpected empty list in %v", ErrInvalidExpression, t)
			}
			tt := expressions[0].Type()
			for _, e := range expressions[1:] {
				if e.Type() != tt {
					return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
				}
			}
			return Array{
				node: p.node(t),
				t:    ArrayOf(tt),
				vals: expressions,
			}, nil
		default:
			panic(fmt.Sprintf("unreachable: unexpected separator token %v", t))
		}
	case lexer.OperatorToken:
		switch operators[t.Value] {
		case equalOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if op.left.Type() != op.right.Type() {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return Equal(op), nil
		case inOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if !isArrayType(op.right.Type()) || op.left.Type() != arrayItemType(op.right.Type()) {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return in(op), nil
		case greaterOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if op.left.Type() != op.right.Type() {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return Greater(op), nil
		case greaterOrEqualOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if op.left.Type() != op.right.Type() {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return GreaterOrEqual(op), nil
		case lessOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if op.left.Type() != op.right.Type() {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return Less(op), nil
		case lessOrEqualOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if op.left.Type() != op.right.Type() {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return LessOrEqual(op), nil
		case andOp:
			op, err := p.parseVariadic(l, t)
			if err != nil {
				return nil, err
			}
			for _, arg := range op.args {
				if arg.Type() != BoolType {
					return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
				}
			}
			return And(op), nil
		case orOp:
			op, err := p.parseVariadic(l, t)
			if err != nil {
				return nil, err
			}
			for _, arg := range op.args {
				if arg.Type() != BoolType {
					return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
				}
			}
			return Or(op), nil
		case notOp:
			op, err := p.parseUnary(l, t)
			if err != nil {
				return nil, err
			}
			if op.arg.Type() != BoolType {
				return nil, fmt.Errorf("%w: type mismatch in %v", ErrInvalidExpression, t)
			}
			return Not(op), nil
		case likeOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if _, ok := op.left.(Column); !ok {
				return nil, fmt.Errorf("%w: unexpected type %v in %v", ErrInvalidExpression, op.left, t)
			}
			if _, ok := op.right.(String); !ok {
				return nil, fmt.Errorf("%w: unexpected type %v in %v", ErrInvalidExpression, op.right, t)
			}
			return Like(op), nil
		case aLikeOp:
			op, err := p.parseBinary(l, t)
			if err != nil {
				return nil, err
			}
			if c, ok := op.left.(Column); !ok && c.Type() != ArrayOf(StringType) {
				return nil, fmt.Errorf("%w: unexpected type %v in %v", ErrInvalidExpression, op.left, t)
			}
			if _, ok := op.right.(String); !ok {
				return nil, fmt.Errorf("%w: unexpected type %v in %v", ErrInvalidExpression, op.right, t)
			}
			return ALike(op), nil
		case dateOp:
			op, err := p.parseUnary(l, t)
			if err != nil {
				return nil, err
			}
			if s, ok := op.arg.(String); ok {
				d, err := p.dateFactory(s.val)
				if err != nil {
					return nil, fmt.Errorf("%w: failed to parse date %q, %s", ErrInvalidExpression, s.val, err)
				}
				return Date{
					node: p.node(t),
					val:  d,
				}, nil
			}
			return nil, fmt.Errorf("%w: unexpected type %v in %v", ErrInvalidExpression, op.arg, t)
		default:
			panic(fmt.Sprintf("unreachable: unexpected operator token %v", t))
		}
	default:
		panic(fmt.Sprintf("unreachable: unexpected token type %v", t))
	}
}

func (p *Filter) consumeSeparator(l *lexer.Lexer, separator rune) error {
	if !l.Next() {
		return fmt.Errorf("%w: unexpected end of expression", ErrInvalidExpression)
	}
	if t, ok := l.Token().(lexer.SeparatorToken); ok && t.Value == separator {
		return nil
	}
	return fmt.Errorf("%w: unexpected token %v", ErrInvalidExpression, l.Token())
}

func (p *Filter) parseUnary(l *lexer.Lexer, t lexer.Token) (unaryOp, error) {
	if err := p.consumeSeparator(l, openParenSep); err != nil {
		return unaryOp{}, err
	}
	expressions, err := p.parseList(l)
	if err != nil {
		return unaryOp{}, err
	}
	if len(expressions) != 1 {
		return unaryOp{}, fmt.Errorf("%w: unexpected number of expressons in %v", ErrInvalidExpression, t)
	}
	return unaryOp{
		node: p.node(t),
		arg:  expressions[0],
	}, nil
}

func (p *Filter) parseBinary(l *lexer.Lexer, t lexer.Token) (binaryOp, error) {
	if err := p.consumeSeparator(l, openParenSep); err != nil {
		return binaryOp{}, err
	}
	expressions, err := p.parseList(l)
	if err != nil {
		return binaryOp{}, err
	}
	if len(expressions) != 2 {
		return binaryOp{}, fmt.Errorf("%w: unexpected number of expressions in %v", ErrInvalidExpression, t)
	}
	return binaryOp{
		node:  p.node(t),
		left:  expressions[0],
		right: expressions[1],
	}, nil
}

func (p *Filter) parseVariadic(l *lexer.Lexer, t lexer.Token) (variadicOp, error) {
	if err := p.consumeSeparator(l, openParenSep); err != nil {
		return variadicOp{}, err
	}
	expressions, err := p.parseList(l)
	if err != nil {
		return variadicOp{}, err
	}
	if len(expressions) < 2 {
		return variadicOp{}, fmt.Errorf("%w: unexpected number of expressions in %v", ErrInvalidExpression, t)
	}
	return variadicOp{
		node: p.node(t),
		args: expressions,
	}, nil
}

func (p *Filter) parseList(l *lexer.Lexer) ([]Expr, error) {
	var expressions []Expr
	for {
		expression, err := p.parse(l)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expression)
		if !l.Next() {
			return nil, fmt.Errorf("%w: unexpected end of expression", ErrInvalidExpression)
		}
		if t, ok := l.Token().(lexer.SeparatorToken); ok {
			if t.Value == closeParenSep {
				return expressions, nil
			}
			if t.Value == commaSep {
				continue
			}
		}
		return nil, fmt.Errorf("%w: unexpected token type %v", ErrInvalidExpression, l.Token())
	}
}
