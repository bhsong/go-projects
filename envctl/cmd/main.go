package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/bhsong/go-projects/envctl/internal/env"
)

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stdout)
		os.Exit(0)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	var err error

	switch subcommand {
	case "parse":
		err = runParse(os.Stdout, args)
	case "check":
		err = runCheck(os.Stdout, args)
	case "merge":
		err = runMerge(os.Stdout, args)
	case "exec":
		err = runExec(os.Stdout, args)
	default:
		printUsage(os.Stdout)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %w\n", err)
		os.Exit(1)
	}
}

func runParse(w io.Writer, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("사용법: envctl parse <파일>")
	}

	path := args[0]

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("runParse: %w", err)
	}
	defer f.Close()

	entries, err := env.Parse(f)
	if err != nil {
		return fmt.Errorf("runParse: %w", err)
	}

	envMap, _ := env.ToMap(entries)

	expMap := make(map[string]string, len(envMap))
	for k, v := range envMap {
		expMap[k] = env.Expand(v, envMap)
	}

	keys := make([]string, 0, len(expMap))
	for k := range expMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(w, "%-20s = %s\n", k, expMap[k])
	}

	return nil
}

func runCheck(w io.Writer, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("사용법: envctl check <파일> [파일...]")
	}

	for _, path := range args {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("runCheck: %w", err)
		}

		entries, err := env.Parse(f)
		if err != nil {
			return fmt.Errorf("runCheck: %w", err)
		}

		_, duplicates := env.ToMap(entries)

		if len(duplicates) == 0 {
			fmt.Fprintf(w, "[%s] 중복 없음 \n", path)
		}

		if len(duplicates) > 0 {
			fmt.Fprintf(w, "[%s]\n", path)
			for _, d := range duplicates {
				lineStars := make([]string, len(d.Lines))
				for i, ln := range d.Lines {
					lineStars[i] = fmt.Sprintf("%d", ln)
				}
				fmt.Fprintf(w, " ⚠  %s: %s번째 줄에서 중복 정의됨\n", d.Key, strings.Join(lineStars, ", "))
			}
		}
	}
	return nil
}
