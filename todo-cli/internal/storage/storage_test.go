package storage

import (
	"os"
	"testing"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

func TestLoad_FileNotExist(t *testing.T) {
	// 없는 파일 경로 -> 에러 없이 빈 슬라이스 반환
	tasks, err := Load("nonexistent.json")

	if err != nil {
		t.Fatalf("파일 없을 때 에러 반환하면 안 됨: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("빈 슬라이스 기대, 실제 길이: %d", len(tasks))
	}
}

func TestSave_AndLoad(t *testing.T) {
	path := "test_tasks.json"
	defer os.Remove(path) // 테스트 후 파일 정리
	original := []task.Task{
		{ID: 1, Title: "테스트 할 일", Done: false},
	}
	if err := Save(path, original); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(loaded))
	}
	if loaded[0].Title != "테스트 할 일" {
		t.Errorf("Title 불일치: %s", loaded[0].Title)
	}
}
