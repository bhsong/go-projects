package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

type JSONStorage struct {
	path string
}

func NewJSONStorage(path string) *JSONStorage {
	return &JSONStorage{path: path}
}

func (j *JSONStorage) Load() ([]task.Task, error) {
	data, err := os.ReadFile(j.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []task.Task{}, nil
		}
		return nil, fmt.Errorf("storage.Load: 파일 읽기 실패: %w", err)
	}
	tasks := []task.Task{}
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, fmt.Errorf("storage.Load: JSON 파싱 실패: %w", err)
	}
	return tasks, nil
}

func (j *JSONStorage) Save(tasks []task.Task) error {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return fmt.Errorf("storage.Save: JSON 변환 실패: %w", err)
	}
	tmp := j.path + ".tmp"
	err = os.WriteFile(tmp, data, 0600)
	if err != nil {
		return fmt.Errorf("storage.Save: 파일 쓰기 실패: %w", err)
	}
	err = os.Rename(tmp, j.path)
	if err != nil {
		err = os.Remove(tmp)
		if err != nil {
			return fmt.Errorf("storage.Save: 임시 파일 삭제 실패: %w", err)
		}
		return fmt.Errorf("storage.Save: 파일 교체 실패: %w", err)
	}
	return nil
}

func (j *JSONStorage) String() string {
	return "JSON 파일: " + j.path
}

var _ task.Storage = (*JSONStorage)(nil)
