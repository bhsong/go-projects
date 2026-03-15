package calc

import (
	"errors"
	"testing"
)

func TestLexer_nextToken(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTokens []Token
	}{
		{
			name:  `[정상] 정수 단독 "42"`,
			input: "42",
			wantTokens: []Token{
				{Type: TokenNumber, Value: "42", Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 소수점 "3.14"`,
			input: "3.14",
			wantTokens: []Token{
				{Type: TokenNumber, Value: "3.14", Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 연산자 "+"`,
			input: "+",
			wantTokens: []Token{
				{Type: TokenPlus, Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 연산자 "-"`,
			input: "-",
			wantTokens: []Token{
				{Type: TokenMinus, Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 연산자 "*"`,
			input: "*",
			wantTokens: []Token{
				{Type: TokenStar, Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 연산자 "/"`,
			input: "/",
			wantTokens: []Token{
				{Type: TokenSlash, Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 괄호 "("`,
			input: "(",
			wantTokens: []Token{
				{Type: TokenLParen, Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 괄호 ")"`,
			input: ")",
			wantTokens: []Token{
				{Type: TokenRParen, Pos: 0},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 공백 무시 "  3  "`,
			input: "  3  ",
			wantTokens: []Token{
				{Type: TokenNumber, Value: "3", Pos: 2},
				{Type: TokenEOF, Pos: 0},
			},
		},
		{
			name:  `[정상] 빈 문자열 ""`,
			input: "",
			wantTokens: []Token{
				{Type: TokenEOF, Pos: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newLexer(tt.input)
			for i, want := range tt.wantTokens {
				got := l.nextToken()
				if got.Type != want.Type {
					t.Errorf("token[%d]: type got=%v, want=%v", i, got.Type, want.Type)
				}
				if want.Value != "" && got.Value != want.Value {
					t.Errorf("token[%d]: value got=%q, want=%q", i, got.Value, want.Value)
				}
				if got.Pos != want.Pos {
					t.Errorf("token[%d]: Pos got=%d, want=%d", i, got.Pos, want.Pos)
				}
			}
		})
	}

	// recover 패턴이 필요해서 테이블 밖에서 별도 처리
	t.Run(`[엣지] 알 수 없는 문자 "@"`, func(t *testing.T) {
		l := newLexer("@")
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("panic이 발생해야 하는데 발생하지 않았습니다")
			}
			var pe *ParseError
			if !errors.As(r.(error), &pe) {
				t.Fatalf("recover 값이 *ParseError가 아닙니다: %T", r)
			}
			if pe.Pos != 0 {
				t.Errorf("ParseError.Pos got=%d, want=0", pe.Pos)
			}
		}()
		l.nextToken()
	})

	t.Run(`[정상] Pos 검증 "1 + 2"`, func(t *testing.T) {
		l := newLexer("1 + 2")
		_ = l.nextToken()     // "1" at Pos 0
		plus := l.nextToken() // "+" at Pos 2 (공백 이후)
		if plus.Type != TokenPlus {
			t.Fatalf("token type got=%v, want=TokenPlus", plus.Type)
		}
		if plus.Pos != 2 {
			t.Errorf("'+' Pos got=%d, want=2", plus.Pos)
		}
	})
}

func TestEval(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		want         float64
		wantErrIs    error // non-nil → errors.Is 검증
		wantParseErr bool  // true → errors.As(*ParseError) 검증
		wantPos      int   // -1 = Pos 미검증; >= 0 = ParseError.Pos 검증
	}{
		{name: `[정상] "3 + 4"`, input: "3 + 4", want: 7, wantPos: -1},
		{name: `[정상] "3 + 4 * 2" (곱셈 우선)`, input: "3 + 4 * 2", want: 11, wantPos: -1},
		{name: `[정상] "10 - 2 * 3"`, input: "10 - 2 * 3", want: 4, wantPos: -1},
		{name: `[정상] "(3 + 4) * 2"`, input: "(3 + 4) * 2", want: 14, wantPos: -1},
		{name: `[정상] "((2 + 3)) * 4" (중첩 괄호)`, input: "((2 + 3)) * 4", want: 20, wantPos: -1},
		{name: `[정상] "1.5 + 2.5" (소수점)`, input: "1.5 + 2.5", want: 4.0, wantPos: -1},
		{name: `[정상] "8 / 2 / 2" (연속 나눗셈)`, input: "8 / 2 / 2", want: 2, wantPos: -1},
		{name: `[정상] "(10 - 2) / 4 + 1 * 3" (복합 수식)`, input: "(10 - 2) / 4 + 1 * 3", want: 5, wantPos: -1},
		{name: `[엣지] "" (빈 입력)`, input: "", wantErrIs: ErrEmptyExpression, wantPos: -1},
		{name: `[엣지] "1 / 0" (0으로 나누기)`, input: "1 / 0", wantErrIs: ErrDivisionByZero, wantPos: -1},
		{name: `[엣지] "1 @" (알 수 없는 문자)`, input: "1 @", wantParseErr: true, wantPos: 2},
		{name: `[엣지] "1 +" (피연산자 없음)`, input: "1 +", wantParseErr: true, wantPos: -1},
		{name: `[엣지] "(1 + 2" (닫는 괄호 없음)`, input: "(1 + 2", wantParseErr: true, wantPos: -1},
		{name: `[엣지] "1 2" (처리되지 않은 토큰)`, input: "1 2", wantParseErr: true, wantPos: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eval(tt.input)

			// 정상 케이스
			if tt.wantErrIs == nil && !tt.wantParseErr {
				if err != nil {
					t.Fatalf("예상치 못한 error: %v", err)
				}
				if got != tt.want {
					t.Errorf("result: got=%v, want=%v", got, tt.want)
				}
				return
			}

			// errors.Is 검증 (ErrEmptyExpression, ErrDivisionByZero)
			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("errors.Is(%v) 실패: got=%v (%T)", tt.wantErrIs, err, err)
				}
				return
			}

			// errors.As 검증 (*ParseError)
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("errors.As(*ParseError) 실패: got=%T %v", err, err)
			}
			if tt.wantPos >= 0 && pe.Pos != tt.wantPos {
				t.Errorf("ParseError.Pos got=%d, want=%d", pe.Pos, tt.wantPos)
			}
		})
	}
}
