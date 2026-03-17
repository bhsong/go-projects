package env

import "fmt"

func ToMap(entries []Entry) (map[string]string, []DuplicateEntry) {
	result := make(map[string]string)
	lineTracker := make(map[string][]int)

	for _, e := range entries {
		lineTracker[e.Key] = append(lineTracker[e.Key], e.Line)
		result[e.Key] = e.Value // 중복이면 나중 값으로 덮어씀
	}

	duplicates := []DuplicateEntry{}

	for key, lines := range lineTracker {
		if len(lines) > 1 {
			// 중복된 키가 있다면 DuplicateEntry로 기록
			duplicates = append(duplicates, DuplicateEntry{
				Key:   key,
				Lines: lines,
			})
		}
	}

	return result, duplicates
}

func MapToSlice(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	return result
}
