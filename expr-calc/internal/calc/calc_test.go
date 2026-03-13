package calc

import (
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
				{Type: TokenNumber, Value: "42"},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 소수점 "3.14"`,
			input: "3.14",
			wantTokens: []Token{
				{Type: TokenNumber, Value: "3.14"},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 연산자 "+"`,
			input: "+",
			wantTokens: []Token{
				{Type: TokenPlus},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 연산자 "-"`,
			input: "-",
			wantTokens: []Token{
				{Type: TokenMinus},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 연산자 "*"`,
			input: "*",
			wantTokens: []Token{
				{Type: TokenStar},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 연산자 "/"`,
			input: "/",
			wantTokens: []Token{
				{Type: TokenSlash},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 괄호 "()"`,
			input: "()",
			wantTokens: []Token{
				{Type: TokenLParen},
				{Type: TokenRParen},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 공백 무시 "  3  "`,
			input: "  3  ",
			wantTokens: []Token{
				{Type: TokenNumber, Value: "3"},
				{Type: TokenEOF},
			},
		},
		{
			name:  `[정상] 빈 문자열 ""`,
			input: "",
			wantTokens: []Token{
				{Type: TokenEOF},
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
			}
		})
	}

	// panic 케이스는 테이블 밖에서 별도 처리 — recover 패턴이 필요해서
	t.Run(`[엣지] 알 수 없는 문자 "@"`, func(t *testing.T) {
		l := newLexer("@")
		defer func() {
			r := recover()
			if r == nil {
				t.Error("panic이 발생해야 하는데 발생하지 않았습니다")
			}
		}()
		l.nextToken()
	})
}

func TestEval(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{name: `[정상] "1 + 2"`, input: "1 + 2", want: 3},
		{name: `[정상] "5 - 3"`, input: "5 - 3", want: 2},
		{name: `[정상] "4 * 3"`, input: "4 * 3", want: 12},
		{name: `[정상] "8 / 2"`, input: "8 / 2", want: 4},
		{name: `[정상] "3 + 4 * 2" (곱셈 우선)`, input: "3 + 4 * 2", want: 11},
		{name: `[정상] "10 - 2 * 3"`, input: "10 - 2 * 3", want: 4},
		{name: `[정상] "(3 + 4) * 2" (괄호 우선순위)`, input: "(3 + 4) * 2", want: 14},
		{name: `[정상] "((2 + 3)) * 4" (중첩 괄호)`, input: "((2 + 3)) * 4", want: 20},
		{name: `[정상] "1.5 + 2.5" (소수점)`, input: "1.5 + 2.5", want: 4.0},
		{name: `[정상] "8 / 2 / 2" (연속 나눗셈)`, input: "8 / 2 / 2", want: 2},
		{name: `[정상] "(10 - 2) / 4 + 1 * 3" (복합 수식)`, input: "(10 - 2) / 4 + 1 * 3", want: 5},
		{name: `[엣지] "" (빈 입력)`, input: "", wantErr: true},
		{name: `[엣지] "1 / 0" (0으로 나누기)`, input: "1 / 0", wantErr: true},
		{name: `[엣지] "3 + * 2" (잘못된 표현식)`, input: "3 + * 2", wantErr: true},
		{name: `[엣지] "(3 + 4" (닫는 괄호 없음)`, input: "(3 + 4", wantErr: true},
		{name: `[엣지] "3 @ 4" (알 수 없는 문자)`, input: "3 @ 4", wantErr: true},
		{name: `[엣지] "3 3" (처리되지 않은 토큰)`, input: "3 3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eval(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error를 기대했으나 nil 반환, result=%v", got)
				}
				return
			}
			if err != nil {
				t.Errorf("예상치 못한 error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("result: got=%v, want=%v", got, tt.want)
			}
		})
	}
}
