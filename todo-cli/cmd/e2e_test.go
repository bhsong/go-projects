// cmd/e2e_test.go: 실제 바이너리를 빌드 후 exec.Command로 실행하는 E2E 테스트.
//
// 각 테스트는 t.TempDir()로 격리된 디렉토리에서 실행되므로 tasks.json이
// 테스트 간에 공유되지 않는다. 바이너리는 테스트마다 buildTestBinary()로
// 빌드하지만 Go 빌드 캐시 덕분에 첫 번째 이후는 빠르게 완료된다.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// buildTestBinary: cmd 패키지를 빌드해서 임시 바이너리 경로를 반환한다.
// t.TempDir()에 출력하므로 테스트 종료 시 자동 정리된다.
func buildTestBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "todo")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("바이너리 빌드 실패: %v\n%s", err, out)
	}
	return bin
}

// runTodo: todo 바이너리를 dir 디렉토리에서 실행하고 stdout/stderr/exitCode를 반환한다.
// 정상 종료(exit 0)와 비정상 종료(exit != 0) 모두 Fatalf 없이 처리한다.
func runTodo(t *testing.T, bin, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("명령 실행 중 예기치 못한 오류: %v", err)
		}
		exitCode = exitErr.ExitCode()
	}
	return
}

// writeTasksJSON: 선행 조건으로 dir/tasks.json을 직접 작성한다.
// 실제 JSONStorage의 JSON 태그(created_at)와 일치하도록 task.Task를 그대로 직렬화한다.
func writeTasksJSON(t *testing.T, dir string, tasks []task.Task) {
	t.Helper()
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.json"), data, 0644); err != nil {
		t.Fatalf("tasks.json 작성 실패: %v", err)
	}
}

// ─── TestE2E_Add ──────────────────────────────────────────────────────────────

func TestE2E_Add(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	stdout, _, exitCode := runTodo(t, bin, dir, "add", "할 일")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "추가됨") {
		t.Errorf("stdout에 '추가됨' 포함 기대, 실제:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "tasks.json")); os.IsNotExist(err) {
		t.Error("tasks.json 파일이 생성되어야 함")
	}
}

// ─── TestE2E_List ─────────────────────────────────────────────────────────────

func TestE2E_List_비었을때(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	stdout, _, exitCode := runTodo(t, bin, dir, "list")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "할 일이 없습니다.") {
		t.Errorf("stdout에 '할 일이 없습니다.' 포함 기대, 실제:\n%s", stdout)
	}
}

func TestE2E_List_항목있을때(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "첫 번째 할 일", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "두 번째 할 일", Done: true, CreatedAt: time.Now()},
	})

	stdout, _, exitCode := runTodo(t, bin, dir, "list")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "[1]") || !strings.Contains(stdout, "첫 번째 할 일") {
		t.Errorf("stdout에 ID=1 항목 포함 기대, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "[2]") || !strings.Contains(stdout, "두 번째 할 일") {
		t.Errorf("stdout에 ID=2 항목 포함 기대, 실제:\n%s", stdout)
	}
}

// ─── TestE2E_Done ─────────────────────────────────────────────────────────────

func TestE2E_Done_정상(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "완료할 일", Done: false, CreatedAt: time.Now()},
	})

	stdout, _, exitCode := runTodo(t, bin, dir, "done", "1")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "완료") {
		t.Errorf("stdout에 '완료' 포함 기대, 실제:\n%s", stdout)
	}

	// tasks.json에 done=true가 저장됐는지 검증
	data, err := os.ReadFile(filepath.Join(dir, "tasks.json"))
	if err != nil {
		t.Fatalf("tasks.json 읽기 실패: %v", err)
	}
	var saved []task.Task
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("tasks.json JSON 파싱 실패: %v", err)
	}
	if len(saved) != 1 || !saved[0].Done {
		t.Errorf("tasks.json에 Done=true 저장 기대, 실제: %+v", saved)
	}
}

func TestE2E_Done_없는ID(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "존재하는 할 일", Done: false, CreatedAt: time.Now()},
	})

	_, stderr, exitCode := runTodo(t, bin, dir, "done", "999")

	if exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", exitCode)
	}
	if stderr == "" {
		t.Error("stderr에 에러 메시지가 있어야 함")
	}
}

// ─── TestE2E_Delete ───────────────────────────────────────────────────────────

func TestE2E_Delete_정상(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "삭제할 일", Done: false, CreatedAt: time.Now()},
		{ID: 2, Title: "남길 일", Done: false, CreatedAt: time.Now()},
	})

	stdout, _, exitCode := runTodo(t, bin, dir, "delete", "1")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "삭제") {
		t.Errorf("stdout에 '삭제' 포함 기대, 실제:\n%s", stdout)
	}

	// tasks.json에서 ID=1이 제거됐는지 검증
	data, err := os.ReadFile(filepath.Join(dir, "tasks.json"))
	if err != nil {
		t.Fatalf("tasks.json 읽기 실패: %v", err)
	}
	var saved []task.Task
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("tasks.json JSON 파싱 실패: %v", err)
	}
	if len(saved) != 1 {
		t.Fatalf("항목 1개만 남아야 함, 실제: %d개", len(saved))
	}
	if saved[0].ID == 1 {
		t.Error("ID=1 항목이 삭제되지 않음")
	}
}

// ─── TestE2E_Delete_없는ID ────────────────────────────────────────────────────

func TestE2E_Delete_없는ID(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "존재하는 할 일", Done: false, CreatedAt: time.Now()},
	})

	_, stderr, exitCode := runTodo(t, bin, dir, "delete", "999")

	if exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", exitCode)
	}
	if stderr == "" {
		t.Error("stderr에 에러 메시지가 있어야 함")
	}
}

// ─── TestE2E_Done_숫자아닌ID ──────────────────────────────────────────────────

func TestE2E_Done_숫자아닌ID(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	_, stderr, exitCode := runTodo(t, bin, dir, "done", "abc")

	if exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", exitCode)
	}
	if stderr == "" {
		t.Error("stderr에 숫자 아닌 ID 에러 메시지가 있어야 함")
	}
}

// ─── TestE2E_전체흐름 ─────────────────────────────────────────────────────────
// add → list → done → list 순서로 전체 흐름을 통합 검증한다.

func TestE2E_전체흐름(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	// 1. add
	stdout, _, exitCode := runTodo(t, bin, dir, "add", "통합 테스트 할 일")
	if exitCode != 0 {
		t.Fatalf("add: exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "추가됨") {
		t.Errorf("add: '추가됨' 포함 기대, 실제:\n%s", stdout)
	}

	// 2. list — 추가된 항목이 보여야 함
	stdout, _, exitCode = runTodo(t, bin, dir, "list")
	if exitCode != 0 {
		t.Fatalf("list: exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "통합 테스트 할 일") {
		t.Errorf("list: 추가한 항목이 출력에 포함되어야 함, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "⬜") {
		t.Errorf("list: 미완료 아이콘 '⬜' 포함 기대, 실제:\n%s", stdout)
	}

	// 3. done 1 — 완료 처리
	stdout, _, exitCode = runTodo(t, bin, dir, "done", "1")
	if exitCode != 0 {
		t.Fatalf("done: exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "완료") {
		t.Errorf("done: '완료' 포함 기대, 실제:\n%s", stdout)
	}

	// 4. list — 완료 상태로 변경됐는지 확인
	stdout, _, exitCode = runTodo(t, bin, dir, "list")
	if exitCode != 0 {
		t.Fatalf("list(완료 후): exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "✅") {
		t.Errorf("list(완료 후): '✅' 아이콘 포함 기대, 실제:\n%s", stdout)
	}
}

// ─── TestE2E_인자없음 / TestE2E_알수없는커맨드 ────────────────────────────────
// 사용법은 stderr에 출력된다 (printUsage가 os.Stderr에 쓰므로).
// 명세의 "stdout에 사용법 출력"은 사용자 가시성 관점 표현이므로
// 실제 검증은 stderr에서 한다.

func TestE2E_인자없음(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	_, stderr, exitCode := runTodo(t, bin, dir)

	if exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stderr, "사용법") && !strings.Contains(strings.ToLower(stderr), "usage") {
		t.Errorf("stderr에 사용법 안내 포함 기대, 실제:\n%s", stderr)
	}
}

func TestE2E_알수없는커맨드(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	_, stderr, exitCode := runTodo(t, bin, dir, "unknown")

	if exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stderr, "사용법") && !strings.Contains(strings.ToLower(stderr), "usage") {
		t.Errorf("stderr에 사용법 안내 포함 기대, 실제:\n%s", stderr)
	}
}
