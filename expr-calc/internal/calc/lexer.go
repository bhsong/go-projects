package calc

import (
	"fmt"
)

type TokenType int

const (
	TokenNumber TokenType = iota // 0: 숫자
	TokenPlus                    // 1: +
	TokenMinus                   // 2: -
	TokenStar                    // 3: *
	TokenSlash                   // 4: /
	TokenLParen                  // 5: (
	TokenRParen                  // 6: )
	TokenEOF                     // 7:
)

type Token struct {
	Type  TokenType
	Value string // 숫자 토큰일 때만 의미 있음
}

type Lexer struct {
	input []rune // string이 아닌 rune 슬라이스 - 유니코드 안전
	pos   int
}

func newLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
	}
}

func (l *Lexer) nextToken() Token {
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t') {
		l.pos++
	}

	if l.pos >= len(l.input) {
		return Token{
			Type: TokenEOF,
		}
	}

	ch := l.input[l.pos]

	if ch >= '0' && ch <= '9' {
		start := l.pos
		for l.pos < len(l.input) && (l.input[l.pos] >= '0' && l.input[l.pos] <= '9' || l.input[l.pos] == '.') {
			l.pos++
		}
		return Token{
			Type:  TokenNumber,
			Value: string(l.input[start:l.pos]),
		}
	}

	l.pos++
	switch ch {
	case '+':
		return Token{Type: TokenPlus}
	case '-':
		return Token{Type: TokenMinus}
	case '*':
		return Token{Type: TokenStar}
	case '/':
		return Token{Type: TokenSlash}
	case '(':
		return Token{Type: TokenLParen}
	case ')':
		return Token{Type: TokenRParen}
	}
	panic(parsePanic{err: fmt.Errorf("calc.Lexer: 알수 없는 문자: %q", ch)})
}
