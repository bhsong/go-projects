// cmd/main_test.go: cmd/main.go의 헬퍼 함수 단위 테스트.
//
// package main으로 선언해야 비공개 헬퍼 함수(runList, runAdd 등)에 접근 가능.
// os.Args를 조작하는 테스트는 전역 상태를 바꾸므로 t.Parallel() 사용 금지.
package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// redirectStdout: 함수 실행 중 os.Stdout을 파이프로 교체하여 출력을 캡처.
// PHP의 ob_start()/ob_get_clean()와 같은 역할.
// 파이프 생성/복원 실패는 테스트를 즉시 종료(Fatalf).
func redirectStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe 생성 실패: %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	f()

	if err := w.Close(); err != nil {
		t.Fatalf("stdout pipe 닫기 실패: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("stdout 캡처 실패: %v", err)
	}
	return buf.String()
}

// redirectStderr: 함수 실행 중 os.Stderr를 파이프로 교체하여 출력을 캡처.
func redirectStderr(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe 생성 실패: %v", err)
	}
	old := os.Stderr
	os.Stderr = w

	f()

	if err := w.Close(); err != nil {
		t.Fatalf("stderr pipe 닫기 실패: %v", err)
	}
	os.Stderr = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("stderr 캡처 실패: %v", err)
	}
	return buf.String()
}

// withArgs: 테스트 중 os.Args를 임시 교체. defer로 복원 보장.
// os.Args는 전역 상태이므로 이 함수를 쓰는 테스트는 병렬 실행 금지.
func withArgs(args []string, f func()) {
	old := os.Args
	os.Args = args
	defer func() { os.Args = old }()
	f()
}

// ===== runList 테스트 =====

// TestRunList_빈슬라이스: 빈 Task 슬라이스 → "할 일이 없습니다." 메시지 출력
func TestRunList_빈슬라이스(t *testing.T) {
	// runList는 출력 대상(stdout/stderr)이 명세에 명시되지 않았으므로 둘 다 캡처.
	var stdOut, stdErr string
	stdErr = redirectStderr(t, func() {
		stdOut = redirectStdout(t, func() {
			runList([]task.Task{})
		})
	})
	combined := stdOut + stdErr
	if !strings.Contains(combined, "없습니다") {
		t.Errorf("빈 목록 메시지(없습니다)가 출력되어야 함, 실제:\n%s", combined)
	}
}

// TestRunList_미완료항목: Done=false → "⬜" 아이콘, ID, Title 모두 stdout에 포함
func TestRunList_미완료항목(t *testing.T) {
	tasks := []task.Task{
		{ID: 1, Title: "PR 리뷰하기", Done: false, CreatedAt: time.Now()},
	}
	out := redirectStdout(t, func() {
		runList(tasks)
	})
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

// TestRunList_완료항목: Done=true → "✅" 아이콘, ID, Title 모두 stdout에 포함
func TestRunList_완료항목(t *testing.T) {
	tasks := []task.Task{
		{ID: 2, Title: "배포 완료", Done: true, CreatedAt: time.Now()},
	}
	out := redirectStdout(t, func() {
		runList(tasks)
	})
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
	tasks := []task.Task{
		{ID: 1, Title: "완료된 일", Done: true, CreatedAt: time.Now()},
		{ID: 2, Title: "남은 일", Done: false, CreatedAt: time.Now()},
	}
	out := redirectStdout(t, func() {
		runList(tasks)
	})
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

// TestRunList_출력순서: 슬라이스 순서대로 출력되는지 — 순서가 바뀌면 안 됨
func TestRunList_출력순서(t *testing.T) {
	tasks := []task.Task{
		{ID: 1, Title: "첫 번째", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "두 번째", Done: false, CreatedAt: time.Now()},
		{ID: 3, Title: "세 번째", Done: false, CreatedAt: time.Now()},
	}
	out := redirectStdout(t, func() {
		runList(tasks)
	})
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

// TestPrintUsage_stderr출력: 사용법이 stderr에 출력되어야 하며 주요 커맨드를 포함
func TestPrintUsage_stderr출력(t *testing.T) {
	errOut := redirectStderr(t, func() {
		printUsage()
	})
	// 사용법/usage 안내 포함 여부
	if !strings.Contains(errOut, "사용법") && !strings.Contains(strings.ToLower(errOut), "usage") {
		t.Errorf("printUsage는 stderr에 사용법 안내를 출력해야 함, 실제:\n%s", errOut)
	}
	// 4개 커맨드 모두 언급되어야 함
	for _, cmd := range []string{"add", "list", "done", "delete"} {
		if !strings.Contains(errOut, cmd) {
			t.Errorf("printUsage에 커맨드 '%s'가 포함되어야 함, 실제:\n%s", cmd, errOut)
		}
	}
}

// TestPrintUsage_stdout비어있음: printUsage는 stdout에는 아무것도 출력하지 않아야 함
func TestPrintUsage_stdout비어있음(t *testing.T) {
	stdOut := redirectStdout(t, func() {
		// stderr 캡처는 별도로 하지 않아도 됨 — stdout만 검사
		printUsage()
	})
	if stdOut != "" {
		t.Errorf("printUsage는 stdout에 출력하면 안 됨, 실제:\n%s", stdOut)
	}
}

// ===== parseID 테스트 =====

// TestParseID_정상: os.Args[2]가 유효한 숫자 → int로 변환하여 반환
// error path(인자 없음, 숫자 아님)는 os.Exit(1)을 호출하므로 E2E에서 검증.
func TestParseID_정상(t *testing.T) {
	var result int
	withArgs([]string{"todo", "done", "7"}, func() {
		result = parseID("done")
	})
	if result != 7 {
		t.Errorf("parseID = %d, 기대: 7", result)
	}
}

// TestParseID_음수허용: 음수 ID도 Atoi가 파싱하므로 그대로 반환 (유효성 검사는 task 패키지 몫)
func TestParseID_음수허용(t *testing.T) {
	var result int
	withArgs([]string{"todo", "done", "-1"}, func() {
		result = parseID("done")
	})
	if result != -1 {
		t.Errorf("parseID = %d, 기대: -1", result)
	}
}

// ===== runAdd 테스트 =====

// TestRunAdd_정상: os.Args[2]에 제목을 주면 새 Task가 추가된 슬라이스 반환 + 성공 메시지 출력
func TestRunAdd_정상(t *testing.T) {
	var result []task.Task
	out := redirectStdout(t, func() {
		withArgs([]string{"todo", "add", "새 할 일"}, func() {
			result = runAdd([]task.Task{})
		})
	})
	if len(result) != 1 {
		t.Fatalf("새 Task 1개 추가 기대, 실제 길이: %d", len(result))
	}
	if result[0].Title != "새 할 일" {
		t.Errorf("Title 불일치: 기대 '새 할 일', 실제 '%s'", result[0].Title)
	}
	if result[0].Done != false {
		t.Error("새 Task는 Done=false여야 함")
	}
	if result[0].ID != 1 {
		t.Errorf("빈 슬라이스에서 첫 추가 → ID=1 기대, 실제: %d", result[0].ID)
	}
	if !strings.Contains(out, "추가됨") {
		t.Errorf("성공 메시지에 '추가됨' 포함 기대, 실제:\n%s", out)
	}
	// 출력에 제목도 포함되어야 함 (명세: "✅ 추가됨: 할 일")
	if !strings.Contains(out, "새 할 일") {
		t.Errorf("성공 메시지에 제목 포함 기대, 실제:\n%s", out)
	}
}

// TestRunAdd_기존목록에추가: 이미 항목이 있는 목록에 추가하면 ID가 max+1
func TestRunAdd_기존목록에추가(t *testing.T) {
	initial := []task.Task{
		{ID: 3, Title: "기존 항목", Done: false, CreatedAt: time.Now()},
	}
	var result []task.Task
	redirectStdout(t, func() {
		withArgs([]string{"todo", "add", "추가 항목"}, func() {
			result = runAdd(initial)
		})
	})
	if len(result) != 2 {
		t.Fatalf("길이 2 기대, 실제: %d", len(result))
	}
	if result[1].ID != 4 {
		t.Errorf("새 항목 ID=4 기대, 실제: %d", result[1].ID)
	}
}

// ===== runComplete 테스트 =====

// TestRunComplete_정상: 존재하는 ID → Done=true로 변경된 슬라이스 반환 + 완료 메시지
func TestRunComplete_정상(t *testing.T) {
	initial := []task.Task{
		{ID: 1, Title: "할 일", Done: false, CreatedAt: time.Now()},
	}
	var result []task.Task
	out := redirectStdout(t, func() {
		withArgs([]string{"todo", "done", "1"}, func() {
			result = runComplete(initial)
		})
	})
	if len(result) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(result))
	}
	if !result[0].Done {
		t.Error("Done=true여야 함")
	}
	if !strings.Contains(out, "완료") {
		t.Errorf("완료 메시지 기대, 실제:\n%s", out)
	}
	// 완료된 ID가 출력에 포함되어야 함 (명세: "✅ 완료 처리: ID 1")
	if !strings.Contains(out, "1") {
		t.Errorf("완료 메시지에 ID(1) 포함 기대, 실제:\n%s", out)
	}
}

// TestRunComplete_다른항목유지: 완료 처리 시 대상이 아닌 항목은 변경 없음
func TestRunComplete_다른항목유지(t *testing.T) {
	initial := []task.Task{
		{ID: 1, Title: "완료할 것", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "유지할 것", Done: false, CreatedAt: time.Now()},
	}
	var result []task.Task
	redirectStdout(t, func() {
		withArgs([]string{"todo", "done", "1"}, func() {
			result = runComplete(initial)
		})
	})
	if len(result) != 2 {
		t.Fatalf("길이 2 기대, 실제: %d", len(result))
	}
	if !result[0].Done {
		t.Error("ID=1의 Done=true여야 함")
	}
	if result[1].Done {
		t.Error("ID=2의 Done은 변경 없어야 함 (false 유지)")
	}
}

// ===== runDelete 테스트 =====

// TestRunDelete_정상: 존재하는 ID → 해당 항목이 빠진 슬라이스 반환 + 삭제 메시지
func TestRunDelete_정상(t *testing.T) {
	initial := []task.Task{
		{ID: 1, Title: "삭제할 것", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "남길 것", Done: false, CreatedAt: time.Now()},
	}
	var result []task.Task
	out := redirectStdout(t, func() {
		withArgs([]string{"todo", "delete", "1"}, func() {
			result = runDelete(initial)
		})
	})
	if len(result) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(result))
	}
	if result[0].ID != 2 {
		t.Errorf("ID=2 항목만 남아야 함, 실제 ID: %d", result[0].ID)
	}
	if !strings.Contains(out, "삭제") {
		t.Errorf("삭제 메시지 기대, 실제:\n%s", out)
	}
	// 삭제된 ID가 출력에 포함되어야 함 (명세: "🗑️  삭제됨: ID 1")
	if !strings.Contains(out, "1") {
		t.Errorf("삭제 메시지에 ID(1) 포함 기대, 실제:\n%s", out)
	}
}

// TestRunDelete_단일항목삭제: 항목이 1개뿐일 때 삭제 → 빈 슬라이스 반환
func TestRunDelete_단일항목삭제(t *testing.T) {
	initial := []task.Task{
		{ID: 1, Title: "유일한 항목", Done: false, CreatedAt: time.Now()},
	}
	var result []task.Task
	redirectStdout(t, func() {
		withArgs([]string{"todo", "delete", "1"}, func() {
			result = runDelete(initial)
		})
	})
	if len(result) != 0 {
		t.Errorf("빈 슬라이스 기대, 실제 길이: %d", len(result))
	}
}
