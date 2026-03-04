package task

import (
	"testing"
	"time"
)

func TestTaskStruct(t *testing.T) {
	tk := Task{
		ID:        1,
		Title:     "Go 공부하기",
		Done:      false,
		CreatedAt: time.Now(),
	}
	if tk.ID != 1 {
		t.Errorf("ID 1 기대, 실제: %d", tk.ID)
	}
	if tk.Title != "Go 공부하기" {
		t.Errorf("Title 불일치: %s", tk.Title)
	}
	if tk.Done != false {
		t.Error("Done은 기본값 false여야 함")
	}
}
