package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// U-13: 정상 JSON 파일 로드 — Task 목록 반환, error=nil
// 주의: 명세의 JSON 태그는 "created_at"이지만 현재 구현은 "createdat" — CreatedAt 체크 시 Red 예상
func TestLoad_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 명세 기준 JSON 포맷 (created_at)
	content := `[{"id":1,"title":"할 일 1","done":false,"created_at":"2026-03-05T09:00:00Z"}]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("테스트 파일 작성 실패: %v", err)
	}

	tasks, err := Load(path)
	if err != nil {
		t.Fatalf("에러 없어야 함: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(tasks))
	}
	if tasks[0].ID != 1 {
		t.Errorf("ID 1 기대, 실제: %d", tasks[0].ID)
	}
	if tasks[0].Title != "할 일 1" {
		t.Errorf("Title 불일치: %s", tasks[0].Title)
	}
	if tasks[0].Done != false {
		t.Error("Done=false 기대")
	}
	// JSON 태그가 "created_at"이어야 CreatedAt이 올바르게 파싱됨
	if tasks[0].CreatedAt.IsZero() {
		t.Error("CreatedAt이 파싱되어야 함 — JSON 태그 'created_at' 확인 필요")
	}
}

// U-14: 파일이 존재하지 않음 — 빈 슬라이스 반환, error=nil (첫 실행 정상 케이스)
func TestLoad_FileNotExist(t *testing.T) {
	tasks, err := Load("/nonexistent/path/tasks.json")

	if err != nil {
		t.Fatalf("파일 없을 때 에러 반환하면 안 됨: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("빈 슬라이스 기대, 실제 길이: %d", len(tasks))
	}
}

// U-15: 빈 JSON 배열 파일 ([]) — 빈 슬라이스 반환, error=nil
func TestLoad_EmptyArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
		t.Fatalf("테스트 파일 작성 실패: %v", err)
	}

	tasks, err := Load(path)
	if err != nil {
		t.Fatalf("빈 배열 파싱 에러 없어야 함: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("빈 슬라이스 기대, 실제 길이: %d", len(tasks))
	}
}

// U-16: 깨진 JSON 파일 — error 반환
func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	if err := os.WriteFile(path, []byte("{ not valid json [[["), 0644); err != nil {
		t.Fatalf("테스트 파일 작성 실패: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("깨진 JSON -> 에러 반환해야 함")
	}
}

// U-17: 정상 Task 목록 저장 — 파일 생성, JSON 내용 일치
func TestSave_Normal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	tasks := []task.Task{
		{ID: 1, Title: "할 일 1", Done: false, CreatedAt: time.Date(2026, 3, 5, 9, 0, 0, 0, time.UTC)},
	}

	if err := Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("파일이 생성되어야 함")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}
	content := string(data)

	if !contains(content, `"id"`) {
		t.Errorf("JSON에 'id' 필드가 포함되어야 함, 실제:\n%s", content)
	}
	if !contains(content, "할 일 1") {
		t.Errorf("JSON에 title이 포함되어야 함, 실제:\n%s", content)
	}
}

// U-18: 빈 슬라이스 저장 — 파일에 [] 기록
func TestSave_EmptySlice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	if err := Save(path, []task.Task{}); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	// MarshalIndent가 [] 또는 [\n] 형태로 출력할 수 있음
	content := string(data)
	if !contains(content, "[") || !contains(content, "]") {
		t.Errorf("빈 배열 JSON 기대, 실제: %s", content)
	}
}

// U-19: 저장 후 Load 했을 때 내용 동일 — 왕복 직렬화 무결성
func TestSave_LoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	original := []task.Task{
		{ID: 1, Title: "할 일 1", Done: false, CreatedAt: time.Date(2026, 3, 5, 9, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "할 일 2", Done: true, CreatedAt: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}
	if len(loaded) != len(original) {
		t.Fatalf("길이 %d 기대, 실제: %d", len(original), len(loaded))
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
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}
