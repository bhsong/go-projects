package calc

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Parser struct {
	lexer   *Lexer
	current Token
}

func newParser(l *Lexer) *Parser {
	p := &Parser{
		lexer: l,
	}

	p.advance()
	return p
}

func (p *Parser) advance() {
	p.current = p.lexer.nextToken()
}

// expr = term { ("+" | "-") term }
func (p *Parser) parseExpr() float64 {
	left := p.parseTerm()

	for p.current.Type == TokenPlus || p.current.Type == TokenMinus {
		op := p.current.Type
		p.advance()
		right := p.parseTerm()
		if op == TokenPlus {
			left += right
		} else {
			left -= right
		}
	}
	return left
}

// term = factor { ("*" | "/" ) factor }
func (p *Parser) parseTerm() float64 {
	left := p.parseFactor()

	for p.current.Type == TokenStar || p.current.Type == TokenSlash {
		op := p.current.Type
		p.advance()
		right := p.parseFactor()
		if op == TokenStar {
			left *= right
		} else {
			if right == 0 {
				panic(ErrDivisionByZero)
			}
			left /= right
		}
	}
	return left
}

// factor = NUMBER | "(" expr ")"
func (p *Parser) parseFactor() float64 {
	if p.current.Type == TokenNumber {
		val, err := strconv.ParseFloat(p.current.Value, 64)
		if err != nil {
			panic(&ParseError{
				Pos: p.current.Pos,
				Msg: fmt.Sprintf("숫자 변환 실패: %q", p.current.Value),
			})
		}
		p.advance()
		return val
	}

	if p.current.Type == TokenLParen {
		p.advance()
		val := p.parseExpr()
		if p.current.Type != TokenRParen {
			panic(&ParseError{
				Pos: p.current.Pos,
				Msg: fmt.Sprintf("닫는 괄호가 없습니다"),
			})
		}
		p.advance()
		return val
	}

	panic(&ParseError{
		Pos: p.current.Pos,
		Msg: fmt.Sprintf("예상치 못한 토큰: %q", p.current.Value),
	})
}

func timing(name string) func() {
	start := time.Now()
	return func() {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, time.Since(start))
	}
}

func Eval(expr string) (result float64, err error) {
	defer timing("Eval")()

	defer func() {
		if r := recover(); r != nil {
			if errors.Is(r.(error), ErrDivisionByZero) {
				err = ErrDivisionByZero
				return
			} else if pe, ok := r.(*ParseError); ok {
				err = pe
				return
			} else {
				panic(r)
			}
		}
	}()
	if expr == "" {
		return 0, ErrEmptyExpression
	}

	l := newLexer(expr)
	p := newParser(l)
	result = p.parseExpr()

	if p.current.Type != TokenEOF {
		panic(&ParseError{
			Pos: p.current.Pos,
			Msg: fmt.Sprintf("처리되지 않은 토큰: %q", p.current.Value),
		})
	}
	return
}
