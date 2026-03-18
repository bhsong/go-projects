package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
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
		defer f.Close()

		entries, err := env.Parse(f)
		if err != nil {
			return fmt.Errorf("runCheck: %w", err)
		}

		_, duplicates := env.ToMap(entries)

		if len(duplicates) == 0 {
			fmt.Fprintf(w, "[%s] 중복 없음 ✓ \n", path)
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

func runMerge(w io.Writer, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("사용법: envctl merge <기준파일> <오버라이드파일> ")
	}

	base, err := loadEnvFile(args[0])
	if err != nil {
		return fmt.Errorf("runMerge: %w", err)
	}

	merged := base

	for _, path := range args[1:] {
		next, err := loadEnvFile(path)
		if err != nil {
			return fmt.Errorf("runMerge: %w", err)
		}
		merged = env.Merge(merged, next)
	}

	expMap := make(map[string]string, len(merged))
	for k, v := range merged {
		expMap[k] = env.Expand(v, merged)
	}

	keys := make([]string, 0, len(expMap))
	for k := range expMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(w, "%s=%s\n", k, expMap[k])
	}
	return nil
}

func runExec(w io.Writer, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("사용법: envctl exec <파일> [파일...] -- <명령어 [인자...]")
	}
	sepIdx := -1
	for i, a := range args {
		if a == "--" {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 {
		return fmt.Errorf("사용법: envctl exec <파일> [파일...] -- <명령어 [인자...]")
	}

	envFiles := args[:sepIdx]
	cmdArgs := args[sepIdx+1:]
	if len(envFiles) == 0 {
		return fmt.Errorf("적어도 하나의 .env 파일이 필요합니다.")
	}
	if len(cmdArgs) == 0 {
		return fmt.Errorf("실행할 명령어가 없습니다.")
	}

	base, err := loadEnvFile(envFiles[0])
	if err != nil {
		return fmt.Errorf("runExec: %w", err)
	}

	merged := base

	for _, path := range envFiles[1:] {
		next, err := loadEnvFile(path)
		if err != nil {
			return fmt.Errorf("runExec: %w", err)
		}
		merged = env.Merge(merged, next)
	}

	expMap := make(map[string]string, len(merged))

	for k, v := range merged {
		expMap[k] = env.Expand(v, merged)
	}

	cmdName := cmdArgs[0]
	cmd := exec.Command(cmdName, cmdArgs[1:]...)
	cmd.Env = append(os.Environ(), env.MapToSlice(expMap)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("runExec: 명령 실행 실패: %w", err)
	}
	return nil
}

func loadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("env.loadEnvFile: 파일 열기 실패: %w", err)
	}
	defer f.Close()

	entries, err := env.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("env.loadEnvFile: 파싱 실패: %w", err)
	}

	envMap, _ := env.ToMap(entries)

	return envMap, nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `사용법: envctl <명령어> [옵션]
명령어:
  parse  <파일>                              .env 파일 파싱 결과 출력 (변수 치환 포함)
  check  <파일> [파일...]                    중복 키 검출
  merge  <기준파일> <오버라이드파일> [...]    우선순위 병합 결과 출력
  exec   <파일> [파일...] -- <명령어> [인자...]
                                             환경변수 주입 후 명령 실행`)
}
