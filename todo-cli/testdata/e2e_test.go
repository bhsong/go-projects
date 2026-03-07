// e2e_test.go: 바이너리를 빌드해서 실제 CLI 명령어를 실행하고 검증.
//
// 주의: Go 빌드 도구는 "testdata/" 디렉터리를 ./... 패턴에서 제외합니다.
// 이 파일은 반드시 아래 명령어로 별도 실행하세요:
//
//	go test ./testdata/ -v
package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

// runMain: 바이너리 빌드 후 테스트 실행. defer가 os.Exit 전에 실행되도록 분리.
func runMain(m *testing.M) int {
	tmp, err := os.MkdirTemp("", "todo-e2e-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "임시 디렉터리 생성 실패: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "todo")

	// testdata/ 에서 ../cmd/ 는 todo-cli/cmd/ 를 가리킴
	build := exec.Command("go", "build", "-o", binaryPath, "../cmd/")
	if out, buildErr := build.CombinedOutput(); buildErr != nil {
		fmt.Fprintf(os.Stderr, "바이너리 빌드 실패: %v\n%s\n", buildErr, out)
		return 1
	}

	return m.Run()
}

// cmdResult: 실행 결과를 담는 구조체
type cmdResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runBinary: 지정된 디렉터리에서 바이너리를 실행하고 결과를 반환.
// tasks.json이 dir에 상대적으로 생성되도록 Dir을 설정함.
func runBinary(t *testing.T, dir string, args ...string) cmdResult {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir // tasks.json이 이 디렉터리에 생성됨

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return cmdResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitCode,
	}
}

// E-01: todo add "할 일" — tasks.json 없음 → 추가 성공 메시지, 파일 생성
func TestE2E_Add(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "add", "할 일")

	if res.exitCode != 0 {
		t.Fatalf("exit code 0 기대, 실제: %d\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}

	output := res.stdout + res.stderr
	if !strings.Contains(output, "추가됨") || !strings.Contains(output, "할 일") {
		t.Errorf("성공 메시지에 '추가됨'과 '할 일'이 포함되어야 함, 실제:\n%s", output)
	}

	// tasks.json 파일이 생성되어야 함
	if _, err := os.Stat(filepath.Join(dir, "tasks.json")); os.IsNotExist(err) {
		t.Error("tasks.json 파일이 생성되어야 함")
	}
}

// E-02: todo list — tasks.json 없음 → "할 일이 없습니다." 출력
func TestE2E_List_Empty(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "list")

	if res.exitCode != 0 {
		t.Fatalf("exit code 0 기대, 실제: %d\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}

	output := res.stdout + res.stderr
	if !strings.Contains(output, "없습니다") {
		t.Errorf("빈 목록 메시지 기대, 실제:\n%s", output)
	}
}

// E-03: todo list — 항목 2개 저장된 상태 → ID/Title/상태 포함 2줄
func TestE2E_List_WithItems(t *testing.T) {
	dir := t.TempDir()

	// 선행 조건: 2개 항목 추가
	runBinary(t, dir, "add", "PR 리뷰하기")
	runBinary(t, dir, "add", "배포 스크립트 수정")

	res := runBinary(t, dir, "list")

	if res.exitCode != 0 {
		t.Fatalf("exit code 0 기대, 실제: %d\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}

	output := res.stdout + res.stderr
	if !strings.Contains(output, "PR 리뷰하기") {
		t.Errorf("첫 번째 항목 'PR 리뷰하기'가 출력에 포함되어야 함, 실제:\n%s", output)
	}
	if !strings.Contains(output, "배포 스크립트 수정") {
		t.Errorf("두 번째 항목 '배포 스크립트 수정'이 출력에 포함되어야 함, 실제:\n%s", output)
	}
	// ID 번호가 출력에 포함되어야 함
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") {
		t.Errorf("ID 번호(1, 2)가 출력에 포함되어야 함, 실제:\n%s", output)
	}
}

// E-04: todo done 1 — ID=1 항목 존재 → 완료 처리 성공 메시지
func TestE2E_Done(t *testing.T) {
	dir := t.TempDir()
	runBinary(t, dir, "add", "PR 리뷰하기")

	res := runBinary(t, dir, "done", "1")

	if res.exitCode != 0 {
		t.Fatalf("exit code 0 기대, 실제: %d\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}

	output := res.stdout + res.stderr
	if !strings.Contains(output, "완료") || !strings.Contains(output, "1") {
		t.Errorf("완료 처리 메시지에 '완료'와 ID(1)가 포함되어야 함, 실제:\n%s", output)
	}
}

// E-05: todo done 999 — 항목 없음 → 에러 메시지, exit code 1
func TestE2E_Done_NotFound(t *testing.T) {
	dir := t.TempDir()

	res := runBinary(t, dir, "done", "999")

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}

	output := res.stdout + res.stderr
	if output == "" {
		t.Error("에러 메시지가 출력되어야 함")
	}
}

// E-06: todo delete 1 — ID=1 항목 존재 → 삭제 성공 메시지
func TestE2E_Delete(t *testing.T) {
	dir := t.TempDir()
	runBinary(t, dir, "add", "삭제할 항목")

	res := runBinary(t, dir, "delete", "1")

	if res.exitCode != 0 {
		t.Fatalf("exit code 0 기대, 실제: %d\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}

	output := res.stdout + res.stderr
	if !strings.Contains(output, "삭제") || !strings.Contains(output, "1") {
		t.Errorf("삭제 메시지에 '삭제'와 ID(1)가 포함되어야 함, 실제:\n%s", output)
	}

	// 삭제 후 list에서 보이지 않아야 함
	listRes := runBinary(t, dir, "list")
	listOutput := listRes.stdout + listRes.stderr
	if strings.Contains(listOutput, "삭제할 항목") {
		t.Error("삭제된 항목이 list에 나타나면 안 됨")
	}
}

// E-07: todo delete 999 — 항목 없음 → 에러 메시지, exit code 1
func TestE2E_Delete_NotFound(t *testing.T) {
	dir := t.TempDir()

	res := runBinary(t, dir, "delete", "999")

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}

	output := res.stdout + res.stderr
	if output == "" {
		t.Error("에러 메시지가 출력되어야 함")
	}
}

// E-08: todo (인자 없음) — 사용법(usage) 출력, exit code 1
func TestE2E_NoArgs(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir) // 인자 없음

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}

	output := res.stdout + res.stderr
	// "사용법" 또는 "usage" 포함
	if !strings.Contains(strings.ToLower(output), "사용법") && !strings.Contains(strings.ToLower(output), "usage") {
		t.Errorf("사용법 안내가 출력되어야 함, 실제:\n%s", output)
	}
}

// E-09: todo unknown — 알 수 없는 명령 메시지, exit code 1
func TestE2E_UnknownCommand(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "unknown")

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}

	output := res.stdout + res.stderr
	if output == "" {
		t.Error("알 수 없는 명령 메시지가 출력되어야 함")
	}
}

// E-10: todo done abc — ID가 숫자 아님 → 에러 메시지, exit code 1
func TestE2E_Done_InvalidID(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "done", "abc")

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}

	output := res.stdout + res.stderr
	if output == "" {
		t.Error("에러 메시지가 출력되어야 함")
	}
}

// E-12: todo add (제목 없이) — runAdd의 인자 누락 처리 → 사용법 출력, exit 1
func TestE2E_Add_인자없음(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "add") // 제목 없이 add만

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}
	output := res.stdout + res.stderr
	// "사용법" 또는 add 관련 안내가 포함되어야 함
	if !strings.Contains(strings.ToLower(output), "사용법") && !strings.Contains(strings.ToLower(output), "usage") {
		t.Errorf("사용법 안내가 출력되어야 함, 실제:\n%s", output)
	}
}

// E-13: todo done (ID 없이) — parseID의 인자 누락 처리 → 사용법 출력, exit 1
func TestE2E_Done_인자없음(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "done") // ID 없이 done만

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}
	output := res.stdout + res.stderr
	if output == "" {
		t.Error("에러/사용법 메시지가 출력되어야 함")
	}
}

// E-14: todo delete (ID 없이) — parseID의 인자 누락 처리 → 사용법 출력, exit 1
func TestE2E_Delete_인자없음(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "delete") // ID 없이 delete만

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}
	output := res.stdout + res.stderr
	if output == "" {
		t.Error("에러/사용법 메시지가 출력되어야 함")
	}
}

// E-15: todo delete abc — parseID의 숫자 아닌 ID 처리 → 에러 메시지, exit 1
func TestE2E_Delete_숫자아닌ID(t *testing.T) {
	dir := t.TempDir()
	res := runBinary(t, dir, "delete", "abc")

	if res.exitCode != 1 {
		t.Errorf("exit code 1 기대, 실제: %d", res.exitCode)
	}
	output := res.stdout + res.stderr
	if output == "" {
		t.Error("에러 메시지가 출력되어야 함")
	}
}

// E-11: add → list → done → list — 전체 흐름 통합 검증
func TestE2E_FullFlow(t *testing.T) {
	dir := t.TempDir()

	// 1) 항목 추가
	addRes := runBinary(t, dir, "add", "PR 리뷰하기")
	if addRes.exitCode != 0 {
		t.Fatalf("add 실패: exit=%d\n%s", addRes.exitCode, addRes.stdout+addRes.stderr)
	}

	// 2) 목록 확인 — 미완료 상태
	listRes1 := runBinary(t, dir, "list")
	if listRes1.exitCode != 0 {
		t.Fatalf("list 실패: exit=%d\n%s", listRes1.exitCode, listRes1.stdout+listRes1.stderr)
	}
	output1 := listRes1.stdout + listRes1.stderr
	if !strings.Contains(output1, "PR 리뷰하기") {
		t.Errorf("list 출력에 'PR 리뷰하기' 포함되어야 함, 실제:\n%s", output1)
	}

	// 3) 완료 처리
	doneRes := runBinary(t, dir, "done", "1")
	if doneRes.exitCode != 0 {
		t.Fatalf("done 실패: exit=%d\n%s", doneRes.exitCode, doneRes.stdout+doneRes.stderr)
	}

	// 4) 목록 재확인 — 완료 상태로 변경되었는지
	listRes2 := runBinary(t, dir, "list")
	if listRes2.exitCode != 0 {
		t.Fatalf("list(완료 후) 실패: exit=%d\n%s", listRes2.exitCode, listRes2.stdout+listRes2.stderr)
	}
	output2 := listRes2.stdout + listRes2.stderr
	if !strings.Contains(output2, "PR 리뷰하기") {
		t.Errorf("완료 후 list 출력에 'PR 리뷰하기' 포함되어야 함, 실제:\n%s", output2)
	}
	// 완료 표시(✅)가 있어야 함
	if !strings.Contains(output2, "✅") {
		t.Errorf("완료 항목에 ✅ 표시가 있어야 함, 실제:\n%s", output2)
	}
}
