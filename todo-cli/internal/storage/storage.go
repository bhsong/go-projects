package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

func Load(path string) ([]task.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage.Load: 파일 읽기 실패: %w", err)
	}
	tasks := []task.Task{}
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, fmt.Errorf("storage.Load: JSON 파싱 실패: %w", err)
	}
	return tasks, err
}

func Save(path string, tasks []task.Task) error {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return fmt.Errorf("storage.Save: JSON 변환 실패: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("storage.Save 파일 쓰기 실패: %w", err)
	}
	return nil
}
