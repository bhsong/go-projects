package task

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// ─── TestAdd ─────────────────────────────────────────────────────────────────
// Add(tasks, title, priority) 세 인자 시그니처 검증.
// priority 유효성 검사와 Task.Priority 필드 저장을 확인한다.

func TestAdd(t *testing.T) {
	t.Run("priority=high 추가 → Priority==\"high\"", func(t *testing.T) {
		tasks, err := Add(nil, "긴급 작업", "high")
		if err != nil {
			t.Fatalf("Add 실패: %v", err)
		}
		if tasks[0].Priority != "high" {
			t.Errorf("Priority \"high\" 기대, 실제: %s", tasks[0].Priority)
		}
	})

	t.Run("priority=normal 추가 → Priority==\"normal\"", func(t *testing.T) {
		tasks, err := Add(nil, "일반 작업", "normal")
		if err != nil {
			t.Fatalf("Add 실패: %v", err)
		}
		if tasks[0].Priority != "normal" {
			t.Errorf("Priority \"normal\" 기대, 실제: %s", tasks[0].Priority)
		}
	})

	t.Run("priority=low 추가 → Priority==\"low\"", func(t *testing.T) {
		tasks, err := Add(nil, "낮은 우선순위", "low")
		if err != nil {
			t.Fatalf("Add 실패: %v", err)
		}
		if tasks[0].Priority != "low" {
			t.Errorf("Priority \"low\" 기대, 실제: %s", tasks[0].Priority)
		}
	})

	t.Run("여러 개 추가 → ID가 순서대로 증가", func(t *testing.T) {
		var tasks []Task
		var err error
		titles := []string{"할 일 1", "할 일 2", "할 일 3"}
		for i, title := range titles {
			tasks, err = Add(tasks, title, "normal")
			if err != nil {
				t.Fatalf("Add(%q) 실패: %v", title, err)
			}
			if tasks[i].ID != i+1 {
				t.Errorf("[%d] ID %d 기대, 실제: %d", i, i+1, tasks[i].ID)
			}
		}
	})

	t.Run("빈 title → 에러 반환", func(t *testing.T) {
		_, err := Add(nil, "", "normal")
		if err == nil {
			t.Error("빈 title -> 에러 반환해야 함")
		}
	})

	t.Run("잘못된 priority(urgent) → 에러 반환", func(t *testing.T) {
		_, err := Add(nil, "작업", "urgent")
		if err == nil {
			t.Error("잘못된 priority -> 에러 반환해야 함")
		}
	})

	t.Run("priority 빈 문자열 → 에러 반환", func(t *testing.T) {
		_, err := Add(nil, "작업", "")
		if err == nil {
			t.Error("빈 priority -> 에러 반환해야 함")
		}
	})
}

// ─── TestFilterTasks ──────────────────────────────────────────────────────────
// FilterTasks(tasks []Task, opts FilterOptions) []Task 검증.
// 원본 슬라이스는 불변이어야 하고, 빈 결과는 nil이 아닌 빈 슬라이스여야 한다.

func TestFilterTasks(t *testing.T) {
	// 픽스처: 미완료-high, 미완료-normal, 완료-high, 완료-low
	fixture := []Task{
		{ID: 1, Title: "미완료-high", Done: false, Priority: "high", CreatedAt: time.Now()},
		{ID: 2, Title: "미완료-normal", Done: false, Priority: "normal", CreatedAt: time.Now()},
		{ID: 3, Title: "완료-high", Done: true, Priority: "high", CreatedAt: time.Now()},
		{ID: 4, Title: "완료-low", Done: true, Priority: "low", CreatedAt: time.Now()},
	}

	t.Run("기본(ShowAll=false, ShowDoneOnly=false) → 미완료만 반환", func(t *testing.T) {
		result := FilterTasks(fixture, FilterOptions{})
		if len(result) != 2 {
			t.Fatalf("미완료 2개 기대, 실제: %d", len(result))
		}
		for _, tk := range result {
			if tk.Done {
				t.Errorf("Done=true 항목이 포함됨: %+v", tk)
			}
		}
	})

	t.Run("ShowAll=true → 완료+미완료 전체 반환", func(t *testing.T) {
		result := FilterTasks(fixture, FilterOptions{ShowAll: true})
		if len(result) != 4 {
			t.Fatalf("전체 4개 기대, 실제: %d", len(result))
		}
	})

	t.Run("ShowDoneOnly=true → 완료만 반환", func(t *testing.T) {
		result := FilterTasks(fixture, FilterOptions{ShowDoneOnly: true})
		if len(result) != 2 {
			t.Fatalf("완료 2개 기대, 실제: %d", len(result))
		}
		for _, tk := range result {
			if !tk.Done {
				t.Errorf("Done=false 항목이 포함됨: %+v", tk)
			}
		}
	})

	t.Run("Priority=high → high이면서 미완료만 반환", func(t *testing.T) {
		result := FilterTasks(fixture, FilterOptions{Priority: "high"})
		if len(result) != 1 {
			t.Fatalf("미완료-high 1개 기대, 실제: %d", len(result))
		}
		if result[0].Priority != "high" {
			t.Errorf("Priority \"high\" 기대, 실제: %s", result[0].Priority)
		}
		if result[0].Done {
			t.Error("Done=false 항목만 기대")
		}
	})

	t.Run("Priority=high + ShowAll=true → high이면서 전체", func(t *testing.T) {
		result := FilterTasks(fixture, FilterOptions{Priority: "high", ShowAll: true})
		if len(result) != 2 {
			t.Fatalf("high 전체 2개 기대, 실제: %d", len(result))
		}
		for _, tk := range result {
			if tk.Priority != "high" {
				t.Errorf("Priority \"high\" 기대, 실제: %s", tk.Priority)
			}
		}
	})

	t.Run("빈 슬라이스 입력 → 빈 슬라이스 반환 (nil 아님)", func(t *testing.T) {
		result := FilterTasks([]Task{}, FilterOptions{})
		if result == nil {
			t.Error("nil이 아닌 빈 슬라이스를 반환해야 함")
		}
		if len(result) != 0 {
			t.Errorf("빈 슬라이스 기대, 실제: %d개", len(result))
		}
	})

	t.Run("조건에 맞는 항목 없음 → 빈 슬라이스 반환 (nil 아님)", func(t *testing.T) {
		// low priority에 Done=false인 항목이 픽스처에 없음
		result := FilterTasks(fixture, FilterOptions{Priority: "low"})
		if result == nil {
			t.Error("nil이 아닌 빈 슬라이스를 반환해야 함")
		}
		if len(result) != 0 {
			t.Errorf("빈 슬라이스 기대, 실제: %d개", len(result))
		}
	})
}

// ─── TestComplete ─────────────────────────────────────────────────────────────

func TestComplete(t *testing.T) {
	t.Run("존재하는 ID — Done=true", func(t *testing.T) {
		s := NewMemoryStorage([]Task{{ID: 1, Title: "테스트", Done: false}})
		tasks, _ := s.Load()

		tasks, err := Complete(tasks, 1)
		if err != nil {
			t.Fatalf("에러 없어야 함: %v", err)
		}
		if err := s.Save(tasks); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		result, _ := s.Load()
		if !result[0].Done {
			t.Error("Done이 true여야 함")
		}
	})

	t.Run("존재하지 않는 ID — 에러 반환", func(t *testing.T) {
		s := NewMemoryStorage([]Task{{ID: 1, Title: "테스트"}})
		tasks, _ := s.Load()

		_, err := Complete(tasks, 999)
		if err == nil {
			t.Fatal("존재하지 않는 ID -> 에러 반환해야 함")
		}
		if !strings.Contains(err.Error(), "999") {
			t.Errorf("에러 메시지에 ID(999) 포함해야 함, 실제: %v", err)
		}
	})

	t.Run("이미 완료된 태스크 — 에러 없이 Done 유지 (멱등성)", func(t *testing.T) {
		s := NewMemoryStorage([]Task{{ID: 1, Title: "이미 완료", Done: true}})
		tasks, _ := s.Load()

		tasks, err := Complete(tasks, 1)
		if err != nil {
			t.Fatalf("이미 완료된 태스크 재완료 -> 에러 없어야 함: %v", err)
		}
		if err := s.Save(tasks); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		result, _ := s.Load()
		if !result[0].Done {
			t.Error("Done=true 유지되어야 함")
		}
	})
}

// ─── TestDelete ───────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	t.Run("존재하는 ID 삭제 — len 감소, 해당 ID 없음", func(t *testing.T) {
		s := NewMemoryStorage([]Task{
			{ID: 1, Title: "삭제할 것"},
			{ID: 2, Title: "남길 것"},
		})
		tasks, _ := s.Load()

		tasks, err := Delete(tasks, 1)
		if err != nil {
			t.Fatalf("에러 없어야 함: %v", err)
		}
		if err := s.Save(tasks); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		result, _ := s.Load()
		if len(result) != 1 {
			t.Fatalf("len 1 기대, 실제: %d", len(result))
		}
		for _, tk := range result {
			if tk.ID == 1 {
				t.Error("삭제된 ID 1이 여전히 존재함")
			}
		}
	})

	t.Run("존재하지 않는 ID — 에러 반환", func(t *testing.T) {
		s := NewMemoryStorage([]Task{{ID: 1, Title: "유일한 항목"}})
		tasks, _ := s.Load()

		_, err := Delete(tasks, 999)
		if err == nil {
			t.Error("존재하지 않는 ID -> 에러 반환해야 함")
		}
	})
}

// ─── TestPrintTasks ───────────────────────────────────────────────────────────
// PrintTasks 출력에 [priority] 레이블이 포함되는지 검증한다.

func TestPrintTasks(t *testing.T) {
	t.Run("미완료 항목 — [priority] 레이블 포함", func(t *testing.T) {
		tasks := []Task{
			{ID: 1, Title: "할 일 1", Done: false, Priority: "high"},
		}
		var buf bytes.Buffer
		PrintTasks(&buf, tasks)

		output := buf.String()
		if !strings.Contains(output, "할 일 1") {
			t.Errorf("출력에 '할 일 1' 포함해야 함, 실제:\n%s", output)
		}
		if !strings.Contains(output, "[high]") {
			t.Errorf("출력에 '[high]' 우선순위 레이블 포함해야 함, 실제:\n%s", output)
		}
		if !strings.Contains(output, "⬜") {
			t.Errorf("미완료 항목은 '⬜' 아이콘이어야 함, 실제:\n%s", output)
		}
	})

	t.Run("완료 항목 — [priority] 레이블 포함", func(t *testing.T) {
		tasks := []Task{
			{ID: 2, Title: "할 일 2", Done: true, Priority: "normal"},
		}
		var buf bytes.Buffer
		PrintTasks(&buf, tasks)

		output := buf.String()
		if !strings.Contains(output, "할 일 2") {
			t.Errorf("출력에 '할 일 2' 포함해야 함, 실제:\n%s", output)
		}
		if !strings.Contains(output, "[normal]") {
			t.Errorf("출력에 '[normal]' 우선순위 레이블 포함해야 함, 실제:\n%s", output)
		}
		if !strings.Contains(output, "✅") {
			t.Errorf("완료 항목은 '✅' 아이콘이어야 함, 실제:\n%s", output)
		}
	})

	t.Run("tasks 비어있음 — '할 일이 없습니다.' 출력", func(t *testing.T) {
		var buf bytes.Buffer
		PrintTasks(&buf, []Task{})

		output := buf.String()
		if !strings.Contains(output, "할 일이 없습니다.") {
			t.Errorf("'할 일이 없습니다.' 포함해야 함, 실제:\n%s", output)
		}
	})
}

// ─── TestDescribeStorage ──────────────────────────────────────────────────────

func TestDescribeStorage(t *testing.T) {
	t.Run("*MemoryStorage — '메모리 (테스트용)' 반환", func(t *testing.T) {
		s := NewMemoryStorage(nil)
		got := DescribeStorage(s)
		if got != "메모리 (테스트용)" {
			t.Errorf("'메모리 (테스트용)' 기대, 실제: %s", got)
		}
	})

	t.Run("nil Storage — '알 수 없는 저장소' 반환", func(t *testing.T) {
		got := DescribeStorage(nil)
		if got != "알 수 없는 저장소" {
			t.Errorf("'알 수 없는 저장소' 기대, 실제: %s", got)
		}
	})
}
