package task

import (
	"fmt"
	"io"
	"strconv"
)

func PrintTasks(w io.Writer, tasks []Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "할 일이 없습니다.")
		return
	}

	for _, t := range tasks {
		if t.Done {
			fmt.Fprintf(w, "✅ ")
		} else {
			fmt.Fprintf(w, "⬜ ")
		}
		fmt.Fprintln(w, "["+strconv.Itoa(t.ID)+"] "+t.Title)
	}

}

func PrintResult(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}

func DescribeStorage(s Storage) string {
	switch v := s.(type) {
	case nil:
		return "알 수 없는 저장소"
	case *MemoryStorage:
		return "메모리 (테스트용)"
	default:
		return fmt.Sprintf("%v", v)
	}
}
