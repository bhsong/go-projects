package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// 컴파일 타임에 JSONStorage가 task.Storage를 구현하는지 확인.
// 이 한 줄이 컴파일되지 않으면 storage 패키지 전체 빌드 실패 → Red.
var _ task.Storage = (*JSONStorage)(nil)

// ─── TestJSONStorageLoad ──────────────────────────────────────────────────────

func TestJSONStorageLoad(t *testing.T) {
	t.Run("파일 있음 — tasks 반환", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "tasks.json")

		content := `[{"id":1,"title":"할 일 1","done":false,"created_at":"2026-03-05T09:00:00Z"}]`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("테스트 파일 작성 실패: %v", err)
		}

		s := NewJSONStorage(path)
		tasks, err := s.Load()
		if err != nil {
			t.Fatalf("에러 없어야 함: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("len 1 기대, 실제: %d", len(tasks))
		}
		if tasks[0].ID != 1 {
			t.Errorf("ID 1 기대, 실제: %d", tasks[0].ID)
		}
		if tasks[0].Title != "할 일 1" {
			t.Errorf("Title 불일치: %s", tasks[0].Title)
		}
		if tasks[0].CreatedAt.IsZero() {
			t.Error("CreatedAt이 파싱되어야 함 — JSON 태그 'created_at' 확인 필요")
		}
	})

	t.Run("파일 없음 — 빈 슬라이스, error nil", func(t *testing.T) {
		s := NewJSONStorage("/nonexistent/path/tasks.json")
		tasks, err := s.Load()

		if err != nil {
			t.Fatalf("파일 없을 때 에러 반환하면 안 됨: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("빈 슬라이스 기대, 실제 len: %d", len(tasks))
		}
	})

	t.Run("JSON 깨진 파일 — error 반환", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "tasks.json")

		if err := os.WriteFile(path, []byte("{ not valid json [[["), 0644); err != nil {
			t.Fatalf("테스트 파일 작성 실패: %v", err)
		}

		s := NewJSONStorage(path)
		_, err := s.Load()
		if err == nil {
			t.Error("깨진 JSON -> 에러 반환해야 함")
		}
	})
}

// ─── TestJSONStorageSave ──────────────────────────────────────────────────────

func TestJSONStorageSave(t *testing.T) {
	t.Run("tasks 저장 후 Load — 동일한 데이터", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "tasks.json")

		original := []task.Task{
			{ID: 1, Title: "할 일 1", Done: false, CreatedAt: time.Date(2026, 3, 5, 9, 0, 0, 0, time.UTC)},
			{ID: 2, Title: "할 일 2", Done: true, CreatedAt: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)},
		}

		s := NewJSONStorage(path)
		if err := s.Save(original); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		loaded, err := s.Load()
		if err != nil {
			t.Fatalf("Load 실패: %v", err)
		}
		if len(loaded) != len(original) {
			t.Fatalf("len %d 기대, 실제: %d", len(original), len(loaded))
		}
		for i := range original {
			if loaded[i].ID != original[i].ID {
				t.Errorf("[%d] ID 불일치: 기대 %d, 실제 %d", i, original[i].ID, loaded[i].ID)
			}
			if loaded[i].Title != original[i].Title {
				t.Errorf("[%d] Title 불일치: 기대 %s, 실제 %s", i, original[i].Title, loaded[i].Title)
			}
			if loaded[i].Done != original[i].Done {
				t.Errorf("[%d] Done 불일치: 기대 %v, 실제 %v", i, original[i].Done, loaded[i].Done)
			}
			if !loaded[i].CreatedAt.Equal(original[i].CreatedAt) {
				t.Errorf("[%d] CreatedAt 불일치: 기대 %v, 실제 %v", i, original[i].CreatedAt, loaded[i].CreatedAt)
			}
		}
	})

	t.Run("atomic write — 임시파일이 최종 경로로 rename됨", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "tasks.json")

		s := NewJSONStorage(path)
		if err := s.Save([]task.Task{{ID: 1, Title: "atomic 테스트"}}); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		// 최종 파일이 존재해야 함
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("최종 파일이 존재해야 함")
		}

		// 임시파일이 남아있지 않아야 함 — dir 내 파일이 정확히 1개
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("디렉토리 읽기 실패: %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("최종 파일 1개만 있어야 함, 실제 파일 수: %d", len(entries))
			for _, e := range entries {
				t.Logf("  남은 파일: %s", e.Name())
			}
		}
	})
}

// ─── TestJSONStorageImplementsStorage ────────────────────────────────────────
// var _ task.Storage = (*JSONStorage)(nil) 선언이 이 파일 최상단에 있으므로
// 컴파일 성공 자체가 interface 구현 여부를 보장한다.

func TestJSONStorageImplementsStorage(t *testing.T) {
	t.Log("JSONStorage가 task.Storage를 구현함 — 컴파일 타임에 확인됨")
}
