package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bhsong/go-projects/expr-calc/internal/calc"
)

func main() {
	var pe *calc.ParseError

	if len(os.Args) < 2 {
		printUsage(os.Stdout)
		os.Exit(0)
	}

	expr := strings.Join(os.Args[1:], " ")

	result, err := calc.Eval(expr)
	if err != nil {
		if errors.Is(err, calc.ErrDivisionByZero) {
			fmt.Fprintf(os.Stderr, "오류: 0으로 나눌 수 없습니다\n")
			os.Exit(1)
		} else if errors.Is(err, calc.ErrEmptyExpression) {
			fmt.Fprintf(os.Stderr, "오류: 빈 표현식입니다\n")
			os.Exit(1)
		} else if errors.As(err, &pe) {
			fmt.Fprintf(os.Stderr, "오류: 위치 %d: %s\n", pe.Pos, pe.Msg)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println(strconv.FormatFloat(result, 'f', -1, 64))
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "사용법: expr-calc <표현식>")
	fmt.Fprintln(w, "예시: ")
	fmt.Fprintln(w, " expr-calc \"3 + 4 * 2\"")
	fmt.Fprintln(w, " expr-calc \"(10 - 2) / 4\"")
}
