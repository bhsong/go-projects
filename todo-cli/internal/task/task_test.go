package task

import (
	"bytes"
	"strings"
	"testing"
)

// ─── TestAdd ─────────────────────────────────────────────────────────────────
// MemoryStorage를 버퍼로 활용해 Add 흐름 전체를 검증한다.
// Add는 이제 ([]Task, error)를 반환해야 하므로 빈 title 검증이 가능해진다.

func TestAdd(t *testing.T) {
	t.Run("빈 storage에 태스크 추가 — len==1, Title 일치", func(t *testing.T) {
		s := NewMemoryStorage(nil)
		tasks, err := s.Load()
		if err != nil {
			t.Fatalf("Load 실패: %v", err)
		}

		tasks, err = Add(tasks, "첫 번째 할 일")
		if err != nil {
			t.Fatalf("Add 실패: %v", err)
		}
		if err := s.Save(tasks); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		result, _ := s.Load()
		if len(result) != 1 {
			t.Fatalf("len 1 기대, 실제: %d", len(result))
		}
		if result[0].Title != "첫 번째 할 일" {
			t.Errorf("Title 불일치: %s", result[0].Title)
		}
	})

	t.Run("여러 개 추가 — ID가 순서대로 증가", func(t *testing.T) {
		s := NewMemoryStorage(nil)
		tasks, _ := s.Load()

		titles := []string{"할 일 1", "할 일 2", "할 일 3"}
		for _, title := range titles {
			var err error
			tasks, err = Add(tasks, title)
			if err != nil {
				t.Fatalf("Add(%q) 실패: %v", title, err)
			}
		}
		if err := s.Save(tasks); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		result, _ := s.Load()
		if len(result) != 3 {
			t.Fatalf("len 3 기대, 실제: %d", len(result))
		}
		for i, tk := range result {
			expected := i + 1
			if tk.ID != expected {
				t.Errorf("[%d] ID %d 기대, 실제: %d", i, expected, tk.ID)
			}
		}
	})

	t.Run("빈 title — 에러 반환", func(t *testing.T) {
		s := NewMemoryStorage(nil)
		tasks, _ := s.Load()

		_, err := Add(tasks, "")
		if err == nil {
			t.Error("빈 title -> 에러 반환해야 함")
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
// io.Writer로 bytes.Buffer를 주입해 출력 내용을 문자열로 검증한다.
// PHP의 ob_start()/ob_get_clean()과 유사한 패턴.

func TestPrintTasks(t *testing.T) {
	t.Run("tasks 있음 — bytes.Buffer에 각 태스크 출력", func(t *testing.T) {
		tasks := []Task{
			{ID: 1, Title: "할 일 1", Done: false},
			{ID: 2, Title: "할 일 2", Done: true},
		}
		var buf bytes.Buffer
		PrintTasks(&buf, tasks)

		output := buf.String()
		if !strings.Contains(output, "할 일 1") {
			t.Errorf("출력에 '할 일 1' 포함해야 함, 실제:\n%s", output)
		}
		if !strings.Contains(output, "할 일 2") {
			t.Errorf("출력에 '할 일 2' 포함해야 함, 실제:\n%s", output)
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
// 타입 스위치 분기를 검증한다.
// 주의: *JSONStorage 케이스는 task ↔ storage 순환 import를 유발하므로
//       task_test.go에서는 테스트하지 않는다 (cmd 레벨에서 통합 테스트로 커버).

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
