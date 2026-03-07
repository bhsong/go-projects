package task

import (
	"strings"
	"testing"
	"time"
)

// U-01: 빈 슬라이스에 첫 항목 추가 — ID=1, Done=false, Title 일치
func TestAdd_FirstItem(t *testing.T) {
	tasks := []Task{}
	before := time.Now()
	tasks = Add(tasks, "첫 번째 할 일")
	after := time.Now()

	if len(tasks) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(tasks))
	}
	if tasks[0].ID != 1 {
		t.Errorf("ID 1 기대, 실제: %d", tasks[0].ID)
	}
	if tasks[0].Done != false {
		t.Error("새 Task는 Done=false여야 함")
	}
	if tasks[0].Title != "첫 번째 할 일" {
		t.Errorf("Title 불일치: %s", tasks[0].Title)
	}
	// Add는 CreatedAt을 time.Now()로 설정해야 함 (함수 명세 참고)
	if tasks[0].CreatedAt.IsZero() {
		t.Error("CreatedAt이 설정되어야 함 (현재 Add에서 미설정)")
	}
	if !tasks[0].CreatedAt.IsZero() {
		if tasks[0].CreatedAt.Before(before) || tasks[0].CreatedAt.After(after) {
			t.Error("CreatedAt이 Add 호출 시각 범위 안이어야 함")
		}
	}
}

// U-02: 기존 최대 ID가 5인 목록에 추가 — 새 항목 ID=6
func TestAdd_AfterMaxID5(t *testing.T) {
	tasks := []Task{{ID: 5, Title: "기존 할 일"}}
	tasks = Add(tasks, "새 할 일")

	if tasks[1].ID != 6 {
		t.Errorf("ID 6 기대, 실제: %d", tasks[1].ID)
	}
}

// U-03: 여러 번 연속 추가 — ID가 1씩 증가
func TestAdd_Sequential(t *testing.T) {
	tasks := []Task{}
	tasks = Add(tasks, "할 일 1")
	tasks = Add(tasks, "할 일 2")
	tasks = Add(tasks, "할 일 3")

	if len(tasks) != 3 {
		t.Fatalf("길이 3 기대, 실제: %d", len(tasks))
	}
	for i, tk := range tasks {
		expected := i + 1
		if tk.ID != expected {
			t.Errorf("인덱스 %d: ID %d 기대, 실제: %d", i, expected, tk.ID)
		}
	}
}

// U-04: 존재하는 ID 완료 처리 — Done=true, error=nil
func TestComplete_Normal(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "테스트", Done: false}}
	updated, err := Complete(tasks, 1)

	if err != nil {
		t.Fatalf("에러 없어야 함: %v", err)
	}
	if !updated[0].Done {
		t.Error("Done이 true여야 함")
	}
}

// U-05: 존재하지 않는 ID — error 반환
func TestComplete_NotFound(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "테스트"}}
	_, err := Complete(tasks, 999)

	if err == nil {
		t.Error("존재하지 않는 ID -> 에러 반환해야 함")
	}
	// 에러 메시지에 ID와 "찾을 수 없음"이 포함되어야 함 (명세: "task.Complete: ID %d 를 찾을 수 없음")
	if err != nil && !strings.Contains(err.Error(), "999") {
		t.Errorf("에러 메시지에 ID(999)가 포함되어야 함, 실제: %v", err)
	}
}

// U-06: 빈 슬라이스에 Complete — error 반환
func TestComplete_EmptySlice(t *testing.T) {
	tasks := []Task{}
	_, err := Complete(tasks, 1)

	if err == nil {
		t.Error("빈 슬라이스에 Complete -> 에러 반환해야 함")
	}
}

// U-07: 이미 Done=true인 항목 Complete 재실행 — Done=true 유지, error=nil (멱등성)
func TestComplete_Idempotent(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "이미 완료", Done: true}}
	updated, err := Complete(tasks, 1)

	if err != nil {
		t.Fatalf("이미 완료된 항목 재완료 -> 에러 없어야 함: %v", err)
	}
	if !updated[0].Done {
		t.Error("Done=true 유지되어야 함")
	}
}

// U-08: 존재하는 ID 삭제 — 해당 항목 제거, 나머지 유지
func TestDelete_Normal(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "삭제할 것"},
		{ID: 2, Title: "남길 것"},
	}
	updated, err := Delete(tasks, 1)

	if err != nil {
		t.Fatalf("에러 없어야 함: %v", err)
	}
	if len(updated) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(updated))
	}
	if updated[0].ID != 2 {
		t.Errorf("ID 2가 남아야 함, 실제: %d", updated[0].ID)
	}
}

// U-09: 목록의 첫 번째 항목 삭제 — 나머지 순서 유지
func TestDelete_FirstItem(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "첫 번째"},
		{ID: 2, Title: "두 번째"},
		{ID: 3, Title: "세 번째"},
	}
	updated, err := Delete(tasks, 1)

	if err != nil {
		t.Fatalf("에러 없어야 함: %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("길이 2 기대, 실제: %d", len(updated))
	}
	if updated[0].ID != 2 || updated[1].ID != 3 {
		t.Errorf("순서 유지 실패: ID=[%d, %d], 기대=[2, 3]", updated[0].ID, updated[1].ID)
	}
}

// U-10: 목록의 마지막 항목 삭제 — 나머지 순서 유지
func TestDelete_LastItem(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "첫 번째"},
		{ID: 2, Title: "두 번째"},
		{ID: 3, Title: "세 번째"},
	}
	updated, err := Delete(tasks, 3)

	if err != nil {
		t.Fatalf("에러 없어야 함: %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("길이 2 기대, 실제: %d", len(updated))
	}
	if updated[0].ID != 1 || updated[1].ID != 2 {
		t.Errorf("순서 유지 실패: ID=[%d, %d], 기대=[1, 2]", updated[0].ID, updated[1].ID)
	}
}

// U-11: 존재하지 않는 ID — error 반환, 원본 슬라이스 변경 없음
func TestDelete_NotFound(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "유일한 항목"}}
	result, err := Delete(tasks, 999)

	if err == nil {
		t.Error("존재하지 않는 ID -> 에러 반환해야 함")
	}
	if len(result) != 1 {
		t.Errorf("원본 변경 없어야 함: 길이 1 기대, 실제: %d", len(result))
	}
	// 에러 메시지에 ID(999)가 포함되어야 함 (명세: "task.Delete: ID %d 를 찾을 수 없음")
	if err != nil && !strings.Contains(err.Error(), "999") {
		t.Errorf("에러 메시지에 ID(999)가 포함되어야 함, 실제: %v", err)
	}
}

// U-12: 빈 슬라이스에 Delete — error 반환
func TestDelete_EmptySlice(t *testing.T) {
	tasks := []Task{}
	_, err := Delete(tasks, 1)

	if err == nil {
		t.Error("빈 슬라이스에 Delete -> 에러 반환해야 함")
	}
}
