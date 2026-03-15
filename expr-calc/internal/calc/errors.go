package calc

import (
	"errors"
	"fmt"
)

var ErrDivisionByZero = errors.New("0으로 나눌 수 없습니다")
var ErrEmptyExpression = errors.New("빈 표현식입니다")

type ParseError struct {
	Pos int    // 문제 발생 위치 (0부터 시작, rune인덱스)
	Msg string // 무슨 문제인가
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("위치 %d: %s", e.Pos, e.Msg)
}
