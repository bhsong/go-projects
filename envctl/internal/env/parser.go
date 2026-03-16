package env

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Entry struct {
	Key   string
	Value string
	Line  int
}

type DuplicateEntry struct {
	Key   string
	Lines []int // 중복 정의된 모든 줄 번호
}

func parseLine(line string, lineNum int) (Entry, bool) {
	line = strings.TrimSpace(line)

	if line == "" || strings.HasPrefix(line, "#") {
		return Entry{}, false
	}

	key, value, found := strings.Cut(line, "=")
	if !found {
		return Entry{}, false
	}

	if key == "" {
		return Entry{}, false
	}

	return Entry{
		Key:   strings.TrimSpace(key),
		Value: strings.TrimSpace(value),
		Line:  lineNum,
	}, true
}

func Parse(r io.Reader) ([]Entry, error) {
	var entries []Entry

	scanner := bufio.NewScanner(r)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if entry, ok := parseLine(scanner.Text(), lineNum); ok {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("env.Parse: 읽기 오류: %w", err)
	}

	return entries, nil
}
