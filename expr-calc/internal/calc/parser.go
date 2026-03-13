package calc

import (
	"fmt"
	"strconv"
)

type parsePanic struct {
	err error
}

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
				panic(parsePanic{err: fmt.Errorf("calc.Parse: 0으로 나눌 수 없습니다.")})
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
			panic(parsePanic{err: fmt.Errorf("calc.Parser: 숫자 변환 실패: %q", p.current.Value)})
		}
		p.advance()
		return val
	}

	if p.current.Type == TokenLParen {
		p.advance()
		val := p.parseExpr()
		if p.current.Type != TokenRParen {
			panic(parsePanic{err: fmt.Errorf("calc.Parser: 닫는 괄호가 없습니다.")})
		}
		p.advance()
		return val
	}

	panic(parsePanic{err: fmt.Errorf("calc.Parser: 예상치 못한 토큰: %q", p.current.Value)})
}

func Eval(expr string) (result float64, err error) {
	//defer timing("Eval")

	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(parsePanic); ok {
				err = pe.err
			} else {
				panic(r)
			}
		}
	}()

	if expr == "" {
		return 0, fmt.Errorf("calc.Eval: 빈 표현식입니다")
	}

	l := newLexer(expr)
	p := newParser(l)
	result = p.parseExpr()

	if p.current.Type != TokenEOF {
		panic(parsePanic{err: fmt.Errorf("calc.Eval: 처리되지 않은 토큰이 있습니다: %q", p.current.Value)})
	}
	return
}
