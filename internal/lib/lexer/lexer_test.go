package lexer

import (
	"errors"
	"reflect"
	"testing"
)

func collect2(l *Lexer) ([]Token, error) {
	var tokens []Token
	for l.Next() {
		tokens = append(tokens, l.Token())
	}
	return tokens, l.Err()
}

func TestLexer(t *testing.T) {
	tests := []struct {
		name      string
		tokenizer *Lexer
		tokens    []Token
		err       error
	}{
		{
			name:      "empty",
			tokenizer: New(nil, nil, ""),
		},
		{
			name:      "number",
			tokenizer: New(nil, nil, "123"),
			tokens: []Token{
				NumberToken{token: token{Pos: 0}, Value: 123},
			},
		},
		{
			name:      "invalid number (leading zero)",
			tokenizer: New(nil, nil, "0123"),
			err:       ErrInvalidNumber,
		},
		{
			name:      "numbers",
			tokenizer: New(nil, nil, "123 456"),
			tokens: []Token{
				NumberToken{token: token{Pos: 0}, Value: 123},
				NumberToken{token: token{Pos: 4}, Value: 456},
			},
		},
		{
			name:      "string",
			tokenizer: New(nil, nil, `"abc"`),
			tokens: []Token{
				StringToken{token: token{Pos: 0}, Value: "abc"},
			},
		},
		{
			name:      "invalid string (unexpected end of input)",
			tokenizer: New(nil, nil, `"abc`),
			err:       ErrInvalidString,
		},
		{
			name:      "escape sequence",
			tokenizer: New(nil, nil, `"a\"\\b"`),
			tokens: []Token{
				StringToken{token: token{Pos: 0}, Value: "a\"\\b"},
			},
		},
		{
			name:      "separator at the end",
			tokenizer: New(nil, []rune{','}, ","),
			tokens: []Token{
				SeparatorToken{token: token{Pos: 0}, Value: ','},
			},
		},
		{
			name:      "separator",
			tokenizer: New(nil, []rune{','}, "1,2,3"),
			tokens: []Token{
				NumberToken{token: token{Pos: 0}, Value: 1},
				SeparatorToken{token: token{Pos: 1}, Value: ','},
				NumberToken{token: token{Pos: 2}, Value: 2},
				SeparatorToken{token: token{Pos: 3}, Value: ','},
				NumberToken{token: token{Pos: 4}, Value: 3},
			},
		},
		{
			name:      "separators",
			tokenizer: New(nil, []rune{',', ':'}, "1,2:: \",::\""),
			tokens: []Token{
				NumberToken{token: token{Pos: 0}, Value: 1},
				SeparatorToken{token: token{Pos: 1}, Value: ','},
				NumberToken{token: token{Pos: 2}, Value: 2},
				SeparatorToken{token: token{Pos: 3}, Value: ':'},
				SeparatorToken{token: token{Pos: 4}, Value: ':'},
				StringToken{token: token{Pos: 6}, Value: ",::"},
			},
		},
		{
			name:      "number to symbol",
			tokenizer: New(nil, nil, "123a"),
			tokens: []Token{
				SymbolToken{token: token{Pos: 0}, Value: "123a"},
			},
		},
		{
			name:      "operator to symbol",
			tokenizer: New([]string{"!=="}, nil, "!=! !="),
			tokens: []Token{
				SymbolToken{token: token{Pos: 0}, Value: "!=!"},
				SymbolToken{token: token{Pos: 4}, Value: "!="},
			},
		},
		{
			name:      "operator overlap",
			tokenizer: New([]string{"as", "assert"}, []rune{','}, "123 as 10 assert,as"),
			tokens: []Token{
				NumberToken{token: token{Pos: 0}, Value: 123},
				OperatorToken{token: token{Pos: 4}, Value: 0},
				NumberToken{token: token{Pos: 7}, Value: 10},
				OperatorToken{token: token{Pos: 10}, Value: 1},
				SeparatorToken{token: token{Pos: 16}, Value: ','},
				OperatorToken{token: token{Pos: 17}, Value: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collect2(tt.tokenizer)
			if !reflect.DeepEqual(got, tt.tokens) {
				t.Errorf("tokenizer.tokenize() = %v, want %v", got, tt.tokens)
			}
			if !errors.Is(err, tt.err) {
				t.Errorf("tokenizer.tokenize() error = %v, want %v", err, tt.err)
			}
		})
	}
}
