// cmd/main_test.go: FlagSet 기반으로 교체된 runXxx 함수의 단위 테스트.
//
// 새 시그니처: runXxx(s task.Storage, w io.Writer, args []string) error
// args는 os.Args[2:] — 서브커맨드 이후의 인자만 전달한다.
// 예: todo add --priority high "할 일" → args = ["--priority", "high", "할 일"]
//
// flag.ContinueOnError를 사용하기 때문에 플래그 파싱 실패 시 os.Exit이 아닌
// error를 반환하므로 테스트에서 에러를 직접 검증할 수 있다.
package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// ===== runAdd 테스트 =====

// TestRunAdd_정상: --priority high 플래그와 제목 → Storage에 Priority=="high" 저장
func TestRunAdd_정상(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	if err := runAdd(s, &buf, []string{"--priority", "high", "할 일"}); err != nil {
		t.Fatalf("runAdd 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Fatalf("Task 1개 추가 기대, 실제: %d", len(tasks))
	}
	if tasks[0].Priority != "high" {
		t.Errorf("Priority \"high\" 기대, 실제: %s", tasks[0].Priority)
	}
	if !strings.Contains(buf.String(), "추가됨") {
		t.Errorf("성공 메시지에 '추가됨' 포함 기대, 실제:\n%s", buf.String())
	}
}

// TestRunAdd_기본priority: --priority 없이 제목만 → Priority=="normal" 저장
func TestRunAdd_기본priority(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	if err := runAdd(s, &buf, []string{"Go 공부하기"}); err != nil {
		t.Fatalf("runAdd 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Fatalf("Task 1개 추가 기대, 실제: %d", len(tasks))
	}
	if tasks[0].Priority != "normal" {
		t.Errorf("Priority \"normal\" 기대, 실제: %s", tasks[0].Priority)
	}
}

// TestRunAdd_잘못된priority: 잘못된 priority 값 → 에러 반환
func TestRunAdd_잘못된priority(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runAdd(s, &buf, []string{"--priority", "urgent", "할 일"})
	if err == nil {
		t.Error("잘못된 priority → 에러 반환해야 함")
	}
}

// TestRunAdd_인자없음: 제목 없이 플래그만 → 에러 반환
func TestRunAdd_인자없음(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runAdd(s, &buf, []string{})
	if err == nil {
		t.Error("인자 없을 때 → 에러 반환해야 함")
	}
}

// ===== runList 테스트 =====

// TestRunList_기본: 미완료+완료 섞인 storage, args=[] → 미완료만 출력
func TestRunList_기본(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "미완료1", Done: false, Priority: "normal", CreatedAt: time.Now()},
		{ID: 2, Title: "미완료2", Done: false, Priority: "high", CreatedAt: time.Now()},
		{ID: 3, Title: "장보기", Done: true, Priority: "normal", CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf, []string{}); err != nil {
		t.Fatalf("runList 실패: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "미완료1") {
		t.Errorf("미완료1이 출력에 포함되어야 함, 실제:\n%s", out)
	}
	if !strings.Contains(out, "미완료2") {
		t.Errorf("미완료2이 출력에 포함되어야 함, 실제:\n%s", out)
	}
	if strings.Contains(out, "장보기") {
		t.Errorf("장보기는 기본 list에 포함되면 안 됨, 실제:\n%s", out)
	}
}

// TestRunList_all: --all 플래그 → 완료+미완료 전체 출력
func TestRunList_all(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "미완료1", Done: false, Priority: "normal", CreatedAt: time.Now()},
		{ID: 2, Title: "장보기", Done: true, Priority: "normal", CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf, []string{"--all"}); err != nil {
		t.Fatalf("runList --all 실패: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "미완료1") {
		t.Errorf("미완료1이 출력에 포함되어야 함, 실제:\n%s", out)
	}
	if !strings.Contains(out, "장보기") {
		t.Errorf("장보기도 출력에 포함되어야 함(--all), 실제:\n%s", out)
	}
}

// TestRunList_done: --done 플래그 → 완료 항목만 출력
func TestRunList_done(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "미완료1", Done: false, Priority: "normal", CreatedAt: time.Now()},
		{ID: 2, Title: "장보기", Done: true, Priority: "normal", CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf, []string{"--done"}); err != nil {
		t.Fatalf("runList --done 실패: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "미완료1") {
		t.Errorf("미완료1은 --done 출력에 포함되면 안 됨, 실제:\n%s", out)
	}
	if !strings.Contains(out, "장보기") {
		t.Errorf("장보기가 --done 출력에 포함되어야 함, 실제:\n%s", out)
	}
}

// TestRunList_priority필터: --priority high → high 항목만 출력
func TestRunList_priority필터(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "긴급", Done: false, Priority: "high", CreatedAt: time.Now()},
		{ID: 2, Title: "일반", Done: false, Priority: "normal", CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf, []string{"--priority", "high"}); err != nil {
		t.Fatalf("runList --priority high 실패: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "긴급") {
		t.Errorf("긴급(high)이 출력에 포함되어야 함, 실제:\n%s", out)
	}
	if strings.Contains(out, "일반") {
		t.Errorf("일반(normal)은 --priority high 출력에 포함되면 안 됨, 실제:\n%s", out)
	}
}

// TestRunList_priority필터_없음: high 항목이 없는 storage에 --priority high → "할 일이 없습니다." 출력
func TestRunList_priority필터_없음(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "일반 작업", Done: false, Priority: "normal", CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runList(s, &buf, []string{"--priority", "high"}); err != nil {
		t.Fatalf("runList --priority high 실패: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "할 일이 없습니다.") {
		t.Errorf("'할 일이 없습니다.' 포함 기대, 실제:\n%s", out)
	}
}

// TestRunList_잘못된priority: 유효하지 않은 priority 값 → 에러 반환
func TestRunList_잘못된priority(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runList(s, &buf, []string{"--priority", "urgent"})
	if err == nil {
		t.Error("잘못된 priority → 에러 반환해야 함")
	}
}

// ===== runDone 테스트 =====

// TestRunDone_정상: args=["1"] → Done=true 확인 (FlagSet 방식)
func TestRunDone_정상(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runDone(s, &buf, []string{"1"}); err != nil {
		t.Fatalf("runDone 실패: %v", err)
	}

	tasks, _ := s.Load()
	if !tasks[0].Done {
		t.Error("Done=true여야 함")
	}
	if !strings.Contains(buf.String(), "완료됨") {
		t.Errorf("완료 메시지 기대, 실제:\n%s", buf.String())
	}
}

// TestRunDone_없는ID: args=["999"] → 에러 반환
func TestRunDone_없는ID(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	err := runDone(s, &buf, []string{"999"})
	if err == nil {
		t.Error("존재하지 않는 ID → 에러 반환해야 함")
	}
}

// TestRunDone_인자없음: args=[] → 에러 반환
func TestRunDone_인자없음(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runDone(s, &buf, []string{})
	if err == nil {
		t.Error("인자 없을 때 → 에러 반환해야 함")
	}
}

// TestRunDone_숫자아닌ID: args=["abc"] → 에러 반환
func TestRunDone_숫자아닌ID(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runDone(s, &buf, []string{"abc"})
	if err == nil {
		t.Error("숫자 아닌 ID → 에러 반환해야 함")
	}
}

// ===== runDelete 테스트 =====

// TestRunDelete_정상: args=["1"] → 삭제 확인 (FlagSet 방식)
func TestRunDelete_정상(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "삭제할 것", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "남길 것", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	if err := runDelete(s, &buf, []string{"1"}); err != nil {
		t.Fatalf("runDelete 실패: %v", err)
	}

	tasks, _ := s.Load()
	if len(tasks) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("ID=2 항목만 남아야 함, 실제 ID: %d", tasks[0].ID)
	}
	if !strings.Contains(buf.String(), "삭제됨") {
		t.Errorf("삭제 메시지 기대, 실제:\n%s", buf.String())
	}
}

// TestRunDelete_없는ID: args=["999"] → 에러 반환
func TestRunDelete_없는ID(t *testing.T) {
	s := task.NewMemoryStorage([]task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	})
	var buf bytes.Buffer

	err := runDelete(s, &buf, []string{"999"})
	if err == nil {
		t.Error("존재하지 않는 ID → 에러 반환해야 함")
	}
}

// TestRunDelete_인자없음: args=[] → 에러 반환
func TestRunDelete_인자없음(t *testing.T) {
	s := task.NewMemoryStorage(nil)
	var buf bytes.Buffer

	err := runDelete(s, &buf, []string{})
	if err == nil {
		t.Error("인자 없을 때 → 에러 반환해야 함")
	}
}

// ===== printUsage 테스트 =====

// TestPrintUsage_사용법출력: 새 플래그 포함 사용법이 출력에 포함되어야 함
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
	// 새 플래그 형식 확인
	if !strings.Contains(out, "--priority") {
		t.Errorf("printUsage에 '--priority' 플래그 안내가 포함되어야 함, 실제:\n%s", out)
	}
}
