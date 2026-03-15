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
	Pos   int    // 토큰의 시작 위치 (디버깅용)
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
			Pos:  0,
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
			Pos:   start,
		}
	}

	l.pos++
	switch ch {
	case '+':
		return Token{Type: TokenPlus, Pos: l.pos - 1}
	case '-':
		return Token{Type: TokenMinus, Pos: l.pos - 1}
	case '*':
		return Token{Type: TokenStar, Pos: l.pos - 1}
	case '/':
		return Token{Type: TokenSlash, Pos: l.pos - 1}
	case '(':
		return Token{Type: TokenLParen, Pos: l.pos - 1}
	case ')':
		return Token{Type: TokenRParen, Pos: l.pos - 1}
	}
	panic(&ParseError{
		Pos: l.pos - 1,
		Msg: fmt.Sprintf("알 수 없는 문자: %q", ch),
	})
}
