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

func TestAdd(t *testing.T) {
	tasks := []Task{}
	tasks = Add(tasks, "첫 번째 할 일")
	if len(tasks) != 1 {
		t.Errorf("길이 1 기대, 실제:%d", len(tasks))
	}
	if tasks[0].ID != 1 {
		t.Errorf("ID 1 기대, 실제: %d", tasks[0].ID)
	}
	if tasks[0].Title != "첫 번째 할 일" {
		t.Errorf("Title 불일치: %s", tasks[0].Title)
	}
	if tasks[0].Done != false {
		t.Error("새 Task는 Done=false여야 함")
	}
}

func TestAdd_IDAutoIncrement(t *testing.T) {
	tasks := []Task{{ID: 5, Title: "기존 할 일"}}
	tasks = Add(tasks, "새 할 일")
	if tasks[1].ID != 6 {
		t.Errorf("ID 6 기대, 실제: %d", tasks[1].ID)
	}
}

func TestComplete(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "테스트", Done: false}}
	updated, err := Complete(tasks, 1)
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if !updated[0].Done {
		t.Errorf("Done이 true여야 함")
	}
}

func TestComplete_NotFound(t *testing.T) {
	tasks := []Task{}
	_, err := Complete(tasks, 999)
	if err == nil {
		t.Error("존재하지 않는 ID -> 에러 반환해야함 ")
	}
}

func TestDelete(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "삭제할 것"},
		{ID: 2, Title: "남 길 것"},
	}
	updated, err := Delete(tasks, 1)
	if err != nil {
		t.Fatalf("에러 발생: %v", err)
	}
	if len(updated) != 1 {
		t.Errorf("길이 1기대, 실제: %d", len(updated))
	}
	if updated[0].ID != 2 {
		t.Errorf("ID 2가 남아야 함. 실제: %d", updated[0].ID)
	}
}

func TestDelete_NotFound(t *testing.T) {
	tasks := []Task{}
	_, err := Delete(tasks, 999)
	if err == nil {
		t.Error("존재하지 않는 ID -> 에러 반환해야 함 ")
	}
}
