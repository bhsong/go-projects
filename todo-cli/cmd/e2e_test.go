// cmd/e2e_test.go: 실제 바이너리를 빌드 후 exec.Command로 실행하는 E2E 테스트.
//
// 각 테스트는 t.TempDir()로 격리된 디렉토리에서 실행되므로 tasks.json이
// 테스트 간에 공유되지 않는다. 바이너리는 buildTestBinary()로 빌드하며
// Go 빌드 캐시 덕분에 첫 번째 이후는 빠르게 완료된다.
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

// readTasksJSON: dir/tasks.json을 읽어 []task.Task로 반환한다.
func readTasksJSON(t *testing.T, dir string) []task.Task {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "tasks.json"))
	if err != nil {
		t.Fatalf("tasks.json 읽기 실패: %v", err)
	}
	var tasks []task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		t.Fatalf("tasks.json JSON 파싱 실패: %v", err)
	}
	return tasks
}

// ─── TestE2E_Add ──────────────────────────────────────────────────────────────

// TestE2E_Add_기본priority: --priority 없이 add → priority=="normal" 저장
func TestE2E_Add_기본priority(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	stdout, _, exitCode := runTodo(t, bin, dir, "add", "Go 공부하기")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "추가됨") {
		t.Errorf("stdout에 '추가됨' 포함 기대, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Go 공부하기") {
		t.Errorf("stdout에 제목 포함 기대, 실제:\n%s", stdout)
	}

	saved := readTasksJSON(t, dir)
	if len(saved) != 1 {
		t.Fatalf("Task 1개 기대, 실제: %d", len(saved))
	}
	if saved[0].Priority != "normal" {
		t.Errorf("Priority \"normal\" 기대, 실제: %s", saved[0].Priority)
	}
}

// TestE2E_Add_highPriority: --priority high 로 add → priority=="high" 저장
func TestE2E_Add_highPriority(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	stdout, _, exitCode := runTodo(t, bin, dir, "add", "--priority", "high", "긴급 작업")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "추가됨") {
		t.Errorf("stdout에 '추가됨' 포함 기대, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "긴급 작업") {
		t.Errorf("stdout에 제목 포함 기대, 실제:\n%s", stdout)
	}

	saved := readTasksJSON(t, dir)
	if len(saved) != 1 {
		t.Fatalf("Task 1개 기대, 실제: %d", len(saved))
	}
	if saved[0].Priority != "high" {
		t.Errorf("Priority \"high\" 기대, 실제: %s", saved[0].Priority)
	}
}

// TestE2E_Add_잘못된priority: --priority urgent → exit code != 0, stderr에 에러
func TestE2E_Add_잘못된priority(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	_, stderr, exitCode := runTodo(t, bin, dir, "add", "--priority", "urgent", "작업")

	if exitCode == 0 {
		t.Error("잘못된 priority → exit code != 0이어야 함")
	}
	if stderr == "" {
		t.Error("stderr에 에러 메시지가 있어야 함")
	}
}

// ─── TestE2E_List ─────────────────────────────────────────────────────────────

// TestE2E_List_기본: 미완료+완료 섞인 상태에서 list → 미완료만 포함
func TestE2E_List_기본(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	// 선행: add 미완료1, 미완료2, 완료1 후 done
	runTodo(t, bin, dir, "add", "미완료1")
	runTodo(t, bin, dir, "add", "미완료2")
	runTodo(t, bin, dir, "add", "완료1")
	runTodo(t, bin, dir, "done", "3")

	stdout, _, exitCode := runTodo(t, bin, dir, "list")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "미완료1") {
		t.Errorf("'미완료1' 포함 기대, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "미완료2") {
		t.Errorf("'미완료2' 포함 기대, 실제:\n%s", stdout)
	}
	if strings.Contains(stdout, "완료1") {
		t.Errorf("'완료1'은 기본 list에 포함되면 안 됨, 실제:\n%s", stdout)
	}
}

// TestE2E_List_all: 동일 선행 후 list --all → 세 항목 모두 포함
func TestE2E_List_all(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	runTodo(t, bin, dir, "add", "미완료1")
	runTodo(t, bin, dir, "add", "미완료2")
	runTodo(t, bin, dir, "add", "완료1")
	runTodo(t, bin, dir, "done", "3")

	stdout, _, exitCode := runTodo(t, bin, dir, "list", "--all")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "미완료1") {
		t.Errorf("'미완료1' 포함 기대, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "미완료2") {
		t.Errorf("'미완료2' 포함 기대, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "완료1") {
		t.Errorf("'완료1' 포함 기대(--all), 실제:\n%s", stdout)
	}
}

// TestE2E_List_done: 동일 선행 후 list --done → 완료1만 포함
func TestE2E_List_done(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	runTodo(t, bin, dir, "add", "미완료1")
	runTodo(t, bin, dir, "add", "미완료2")
	runTodo(t, bin, dir, "add", "완료1")
	runTodo(t, bin, dir, "done", "3")

	stdout, _, exitCode := runTodo(t, bin, dir, "list", "--done")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if strings.Contains(stdout, "미완료1") {
		t.Errorf("'미완료1'은 --done 출력에 포함되면 안 됨, 실제:\n%s", stdout)
	}
	if strings.Contains(stdout, "미완료2") {
		t.Errorf("'미완료2'은 --done 출력에 포함되면 안 됨, 실제:\n%s", stdout)
	}
	if !strings.Contains(stdout, "완료1") {
		t.Errorf("'완료1' 포함 기대(--done), 실제:\n%s", stdout)
	}
}

// TestE2E_List_priority필터: high/normal 추가 후 list --all --priority high → high만 포함
func TestE2E_List_priority필터(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	runTodo(t, bin, dir, "add", "--priority", "high", "긴급")
	runTodo(t, bin, dir, "add", "일반")

	stdout, _, exitCode := runTodo(t, bin, dir, "list", "--all", "--priority", "high")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "긴급") {
		t.Errorf("'긴급'(high) 포함 기대, 실제:\n%s", stdout)
	}
	if strings.Contains(stdout, "일반") {
		t.Errorf("'일반'(normal)은 --priority high 출력에 포함되면 안 됨, 실제:\n%s", stdout)
	}
}

// ─── TestE2E_Done (기존 유지) ──────────────────────────────────────────────────

func TestE2E_Done_정상(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "완료할 일", Done: false, Priority: "normal", CreatedAt: time.Now()},
	})

	stdout, _, exitCode := runTodo(t, bin, dir, "done", "1")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "완료됨") {
		t.Errorf("stdout에 '완료됨' 포함 기대, 실제:\n%s", stdout)
	}

	saved := readTasksJSON(t, dir)
	if len(saved) != 1 || !saved[0].Done {
		t.Errorf("tasks.json에 Done=true 저장 기대, 실제: %+v", saved)
	}
}

// ─── TestE2E_Delete (기존 유지) ───────────────────────────────────────────────

func TestE2E_Delete_정상(t *testing.T) {
	bin := buildTestBinary(t)
	dir := t.TempDir()

	writeTasksJSON(t, dir, []task.Task{
		{ID: 1, Title: "삭제할 일", Done: false, Priority: "normal", CreatedAt: time.Now()},
		{ID: 2, Title: "남길 일", Done: false, Priority: "normal", CreatedAt: time.Now()},
	})

	stdout, _, exitCode := runTodo(t, bin, dir, "delete", "1")

	if exitCode != 0 {
		t.Errorf("exit code 0 기대, 실제: %d", exitCode)
	}
	if !strings.Contains(stdout, "삭제됨") {
		t.Errorf("stdout에 '삭제됨' 포함 기대, 실제:\n%s", stdout)
	}

	saved := readTasksJSON(t, dir)
	if len(saved) != 1 {
		t.Fatalf("항목 1개만 남아야 함, 실제: %d개", len(saved))
	}
	if saved[0].ID == 1 {
		t.Error("ID=1 항목이 삭제되지 않음")
	}
}

// ─── TestE2E_인자없음 / TestE2E_알수없는커맨드 (기존 유지) ─────────────────────

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
