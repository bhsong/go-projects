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
			return []task.Task{}, nil
		}
		return nil, fmt.Errorf("loadTasks: 파일 읽기 실패: %w", err)
	}
	var tasks []task.Task
	err = json.Unmarshal(data, &tasks)
	return tasks, err
}

func Save(path string, t []task.Task) error {
	data, err := json.MarshalIndent(t, "", " ")

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
