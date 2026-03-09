// cmd/main_test.go: cmd/main.go의 헬퍼 함수 단위 테스트.
//
// package main으로 선언해야 비공개 헬퍼 함수(runList, runAdd 등)에 접근 가능.
// runXxx 함수가 io.Writer를 직접 받으므로 bytes.Buffer를 주입해서 출력을 검증한다.
// args 파라미터는 os.Args 포맷: ["todo", "서브커맨드", "인자..."]
// os.Exit(1)을 호출하는 경로(숫자 아닌 ID 등)는 E2E 테스트에서 검증한다.
package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// ===== runList 테스트 =====
// 새 시그니처: runList(s task.Storage, w io.Writer) error
// io.Writer로 bytes.Buffer를 주입해서 출력 내용을 직접 검증한다.

// TestRunList_빈슬라이스: 빈 Storage → "할 일이 없습니다." 출력
func TestRunList_빈슬라이스(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	if err := runList(s, &buf); err != nil {
		t.Fatalf("runList 실패: %v", err)
	}
	if !strings.Contains(buf.String(), "없습니다") {
		t.Errorf("빈 목록 메시지(없습니다)가 출력되어야 함, 실제:\n%s", buf.String())
	}
}

// TestRunList_미완료항목: Done=false → "⬜" 아이콘, ID, Title 포함
func TestRunList_미완료항목(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "PR 리뷰하기", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf); err != nil {
		t.Fatalf("runList 실패: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "⬜") {
		t.Errorf("미완료 항목은 '⬜' 아이콘이어야 함, 실제:\n%s", out)
	}
	if !strings.Contains(out, "[1]") {
		t.Errorf("ID가 '[1]' 형식으로 포함되어야 함, 실제:\n%s", out)
	}
	if !strings.Contains(out, "PR 리뷰하기") {
		t.Errorf("Title이 출력에 포함되어야 함, 실제:\n%s", out)
	}
}

// TestRunList_완료항목: Done=true → "✅" 아이콘, ID, Title 포함
func TestRunList_완료항목(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 2, Title: "배포 완료", Done: true, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf); err != nil {
		t.Fatalf("runList 실패: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "✅") {
		t.Errorf("완료 항목은 '✅' 아이콘이어야 함, 실제:\n%s", out)
	}
	if !strings.Contains(out, "[2]") {
		t.Errorf("ID가 '[2]' 형식으로 포함되어야 함, 실제:\n%s", out)
	}
	if !strings.Contains(out, "배포 완료") {
		t.Errorf("Title이 출력에 포함되어야 함, 실제:\n%s", out)
	}
}

// TestRunList_혼합목록: Done=true/false 혼합 → 각각 올바른 아이콘으로 출력
func TestRunList_혼합목록(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "완료된 일", Done: true, CreatedAt: time.Now()},
		{ID: 2, Title: "남은 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf); err != nil {
		t.Fatalf("runList 실패: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "✅") {
		t.Errorf("완료 항목의 '✅' 아이콘 누락, 실제:\n%s", out)
	}
	if !strings.Contains(out, "⬜") {
		t.Errorf("미완료 항목의 '⬜' 아이콘 누락, 실제:\n%s", out)
	}
	if !strings.Contains(out, "완료된 일") || !strings.Contains(out, "남은 일") {
		t.Errorf("모든 Title이 출력되어야 함, 실제:\n%s", out)
	}
}

// TestRunList_출력순서: 슬라이스 순서대로 출력되는지
func TestRunList_출력순서(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "첫 번째", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "두 번째", Done: false, CreatedAt: time.Now()},
		{ID: 3, Title: "세 번째", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf); err != nil {
		t.Fatalf("runList 실패: %v", err)
	}
	out := buf.String()
	idx1 := strings.Index(out, "첫 번째")
	idx2 := strings.Index(out, "두 번째")
	idx3 := strings.Index(out, "세 번째")
	if idx1 < 0 || idx2 < 0 || idx3 < 0 {
		t.Fatalf("모든 항목이 출력되어야 함, 실제:\n%s", out)
	}
	if !(idx1 < idx2 && idx2 < idx3) {
		t.Errorf("슬라이스 순서대로 출력되어야 함, 실제:\n%s", out)
	}
}

// ===== printUsage 테스트 =====
// 새 시그니처: printUsage(w io.Writer)

// TestPrintUsage_사용법출력: 사용법과 4개 커맨드가 출력에 포함되어야 함
func TestPrintUsage_사용법출력(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	out := buf.String()

	if !strings.Contains(out, "사용법") && !strings.Contains(strings.ToLower(out), "usage") {
		t.Errorf("printUsage는 사용법 안내를 출력해야 함, 실제:\n%s", out)
	}
	for _, cmd := range []string{"add", "list", "done", "delete"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("printUsage에 커맨드 '%s'가 포함되어야 함, 실제:\n%s", cmd, out)
		}
	}
}

// ===== runAdd 테스트 =====
// 새 시그니처: runAdd(s task.Storage, w io.Writer, args []string) error
// args는 os.Args 포맷: args[0]=프로그램명, args[1]="add", args[2]=제목

// TestRunAdd_정상: 새 Task가 Storage에 추가됨 + 성공 메시지 출력
func TestRunAdd_정상(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	if err := runAdd(s, &buf, []string{"todo", "add", "새 할 일"}); err != nil {
		t.Fatalf("runAdd 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Fatalf("새 Task 1개 추가 기대, 실제 길이: %d", len(tasks))
	}
	if tasks[0].Title != "새 할 일" {
		t.Errorf("Title 불일치: 기대 '새 할 일', 실제 '%s'", tasks[0].Title)
	}
	if tasks[0].Done != false {
		t.Error("새 Task는 Done=false여야 함")
	}
	if tasks[0].ID != 1 {
		t.Errorf("빈 Storage에서 첫 추가 → ID=1 기대, 실제: %d", tasks[0].ID)
	}
	out := buf.String()
	if !strings.Contains(out, "추가됨") {
		t.Errorf("성공 메시지에 '추가됨' 포함 기대, 실제:\n%s", out)
	}
	if !strings.Contains(out, "새 할 일") {
		t.Errorf("성공 메시지에 제목 포함 기대, 실제:\n%s", out)
	}
}

// TestRunAdd_기존목록에추가: 이미 항목이 있는 Storage에 추가하면 ID가 max+1
func TestRunAdd_기존목록에추가(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 3, Title: "기존 항목", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runAdd(s, &buf, []string{"todo", "add", "추가 항목"}); err != nil {
		t.Fatalf("runAdd 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 2 {
		t.Fatalf("길이 2 기대, 실제: %d", len(tasks))
	}
	if tasks[1].ID != 4 {
		t.Errorf("새 항목 ID=4 기대, 실제: %d", tasks[1].ID)
	}
}

// TestRunAdd_인자없음: args[2]가 없을 때 → 에러 반환 (os.Exit 없음)
func TestRunAdd_인자없음(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runAdd(s, &buf, []string{"todo", "add"})
	if err == nil {
		t.Error("인자 없을 때 → 에러 반환해야 함")
	}
}

// ===== runDone 테스트 =====
// 새 시그니처: runDone(s task.Storage, w io.Writer, args []string) error
// args는 os.Args 포맷: args[0]=프로그램명, args[1]="done", args[2]=ID

// TestRunDone_정상: 존재하는 ID → Done=true로 변경 + 완료 메시지
func TestRunDone_정상(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runDone(s, &buf, []string{"todo", "done", "1"}); err != nil {
		t.Fatalf("runDone 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(tasks))
	}
	if !tasks[0].Done {
		t.Error("Done=true여야 함")
	}
	out := buf.String()
	if !strings.Contains(out, "완료") {
		t.Errorf("완료 메시지 기대, 실제:\n%s", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("완료 메시지에 ID(1) 포함 기대, 실제:\n%s", out)
	}
}

// TestRunDone_다른항목유지: 완료 처리 시 대상이 아닌 항목은 변경 없음
func TestRunDone_다른항목유지(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "완료할 것", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "유지할 것", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runDone(s, &buf, []string{"todo", "done", "1"}); err != nil {
		t.Fatalf("runDone 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 2 {
		t.Fatalf("길이 2 기대, 실제: %d", len(tasks))
	}
	if !tasks[0].Done {
		t.Error("ID=1의 Done=true여야 함")
	}
	if tasks[1].Done {
		t.Error("ID=2의 Done은 변경 없어야 함 (false 유지)")
	}
}

// TestRunDone_없는ID: 존재하지 않는 ID → 에러 반환 (os.Exit 없음)
func TestRunDone_없는ID(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	err := runDone(s, &buf, []string{"todo", "done", "999"})
	if err == nil {
		t.Error("존재하지 않는 ID → 에러 반환해야 함")
	}
}

// TestRunDone_인자없음: args[2]가 없을 때 → 에러 반환 (os.Exit 없음)
func TestRunDone_인자없음(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runDone(s, &buf, []string{"todo", "done"})
	if err == nil {
		t.Error("인자 없을 때 → 에러 반환해야 함")
	}
}

// TestRunDone_숫자아닌ID: args[2]가 숫자 아님 → 에러 반환 (os.Exit 없음)
func TestRunDone_숫자아닌ID(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runDone(s, &buf, []string{"todo", "done", "abc"})
	if err == nil {
		t.Error("숫자 아닌 ID → 에러 반환해야 함")
	}
}

// ===== runDelete 테스트 =====
// 새 시그니처: runDelete(s task.Storage, w io.Writer, args []string) error
// args는 os.Args 포맷: args[0]=프로그램명, args[1]="delete", args[2]=ID

// TestRunDelete_정상: 존재하는 ID → 해당 항목이 Storage에서 제거 + 삭제 메시지
func TestRunDelete_정상(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "삭제할 것", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "남길 것", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runDelete(s, &buf, []string{"todo", "delete", "1"}); err != nil {
		t.Fatalf("runDelete 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("ID=2 항목만 남아야 함, 실제 ID: %d", tasks[0].ID)
	}
	out := buf.String()
	if !strings.Contains(out, "삭제") {
		t.Errorf("삭제 메시지 기대, 실제:\n%s", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("삭제 메시지에 ID(1) 포함 기대, 실제:\n%s", out)
	}
}

// TestRunDelete_단일항목삭제: 항목이 1개뿐일 때 삭제 → 빈 Storage
func TestRunDelete_단일항목삭제(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "유일한 항목", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runDelete(s, &buf, []string{"todo", "delete", "1"}); err != nil {
		t.Fatalf("runDelete 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 0 {
		t.Errorf("빈 슬라이스 기대, 실제 길이: %d", len(tasks))
	}
}

// TestRunDelete_없는ID: 존재하지 않는 ID → 에러 반환 (os.Exit 없음)
func TestRunDelete_없는ID(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	err := runDelete(s, &buf, []string{"todo", "delete", "999"})
	if err == nil {
		t.Error("존재하지 않는 ID → 에러 반환해야 함")
	}
}

// TestRunDelete_인자없음: args[2]가 없을 때 → 에러 반환 (os.Exit 없음)
func TestRunDelete_인자없음(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runDelete(s, &buf, []string{"todo", "delete"})
	if err == nil {
		t.Error("인자 없을 때 → 에러 반환해야 함")
	}
}
