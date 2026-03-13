package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bhsong/go-projects/expr-calc/internal/calc"
)

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stdout)
		os.Exit(0)
	}

	expr := strings.Join(os.Args[1:], " ")

	result, err := calc.Eval(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(strconv.FormatFloat(result, 'f', -1, 64))
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "사용법: expr-calc ,<표현식>")
	fmt.Fprintln(w, "예시: ")
	fmt.Fprintln(w, " expr-calc \"3 + 4 * 2\"")
	fmt.Fprintln(w, " expr-calc \"(10 - 2) / 4\"")
}
