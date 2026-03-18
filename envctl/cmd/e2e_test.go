package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath는 TestMain에서 빌드된 바이너리 경로를 저장한다.
var binaryPath string

// TestMain은 테스트 실행 전 바이너리를 한 번만 빌드하고,
// 테스트 종료 후 정리한다.
// — 바이너리를 매 테스트마다 빌드하지 않기 위해 TestMain 패턴을 사용.
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "envctl-e2e-*")
	if err != nil {
		panic("임시 디렉토리 생성 실패: " + err.Error())
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "envctl")

	// cmd 디렉토리(패키지 루트)에서 빌드한다.
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		// 빌드 실패도 Red 상태임을 출력하고 비정상 종료.
		// 테스트 파일 자체의 컴파일은 성공해야 Red 판정이 가능하다.
		os.Stderr.Write(out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// writeTempEnv는 content를 임시 디렉토리에 .env 파일로 저장하고 경로를 반환한다.
func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("임시 .env 파일 쓰기 실패: %v", err)
	}
	return path
}

// runEnvctl은 binaryPath를 args로 실행하고 stdout, stderr, ExitCode를 반환한다.
func runEnvctl(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// ─────────────────────────────────────────────────────────────────────────────
// E2E: parse
// ─────────────────────────────────────────────────────────────────────────────

func TestE2E_Parse_정상(t *testing.T) {
	path := writeTempEnv(t, "DB_HOST=localhost\nDB_PORT=5432\n")
	stdout, _, exitCode := runEnvctl(t, "parse", path)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "DB_HOST") {
		t.Errorf("stdout에 DB_HOST가 없음:\n%s", stdout)
	}
	if !strings.Contains(stdout, "localhost") {
		t.Errorf("stdout에 localhost가 없음:\n%s", stdout)
	}
}

func TestE2E_Parse_변수치환(t *testing.T) {
	path := writeTempEnv(t, "HOST=localhost\nURL=$HOST/app\n")
	stdout, _, exitCode := runEnvctl(t, "parse", path)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "localhost/app") {
		t.Errorf("stdout에 localhost/app이 없음 (변수 치환 실패):\n%s", stdout)
	}
}

func TestE2E_Parse_파일없음(t *testing.T) {
	_, stderr, exitCode := runEnvctl(t, "parse", "notexist.env")

	if exitCode == 0 {
		t.Errorf("exit code = 0, want non-zero (파일 없음 에러)")
	}
	if stderr == "" {
		t.Errorf("stderr가 비어 있음, 오류 메시지 기대")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// E2E: check
// ─────────────────────────────────────────────────────────────────────────────

func TestE2E_Check_중복없음(t *testing.T) {
	path := writeTempEnv(t, "KEY=value\nOTHER=123\n")
	stdout, _, exitCode := runEnvctl(t, "check", path)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "중복 없음 ✓") {
		t.Errorf("stdout에 '중복 없음 ✓'가 없음:\n%s", stdout)
	}
}

func TestE2E_Check_중복있음(t *testing.T) {
	path := writeTempEnv(t, "KEY=a\nOTHER=b\nKEY=c\n")
	stdout, _, exitCode := runEnvctl(t, "check", path)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (중복 탐지는 에러 아님)", exitCode)
	}
	if !strings.Contains(stdout, "⚠") {
		t.Errorf("stdout에 ⚠가 없음:\n%s", stdout)
	}
	if !strings.Contains(stdout, "KEY") {
		t.Errorf("stdout에 KEY가 없음:\n%s", stdout)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// E2E: merge
// ─────────────────────────────────────────────────────────────────────────────

func TestE2E_Merge_우선순위(t *testing.T) {
	base := writeTempEnv(t, "A=1\nB=2\n")
	// local은 다른 TempDir에 만들어야 한다.
	localDir := t.TempDir()
	localPath := filepath.Join(localDir, "local.env")
	if err := os.WriteFile(localPath, []byte("B=3\nC=4\n"), 0o600); err != nil {
		t.Fatalf("local.env 쓰기 실패: %v", err)
	}

	stdout, _, exitCode := runEnvctl(t, "merge", base, localPath)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "B=3") {
		t.Errorf("stdout에 B=3이 없음 (override 우선 실패):\n%s", stdout)
	}
	if !strings.Contains(stdout, "A=1") {
		t.Errorf("stdout에 A=1이 없음:\n%s", stdout)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// E2E: exec
// ─────────────────────────────────────────────────────────────────────────────

func TestE2E_Exec_환경변수주입(t *testing.T) {
	path := writeTempEnv(t, "ENVCTL_TEST_VAR=hello\n")
	stdout, _, exitCode := runEnvctl(t, "exec", path, "--", "env")

	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "ENVCTL_TEST_VAR=hello") {
		t.Errorf("stdout에 ENVCTL_TEST_VAR=hello가 없음:\n%s", stdout)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// E2E: 기타
// ─────────────────────────────────────────────────────────────────────────────

func TestE2E_인자없음(t *testing.T) {
	stdout, _, exitCode := runEnvctl(t)

	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (인자 없이 사용법 출력)", exitCode)
	}
	if stdout == "" {
		t.Errorf("stdout이 비어 있음, 사용법 출력 기대")
	}
}

func TestE2E_알수없는커맨드(t *testing.T) {
	_, _, exitCode := runEnvctl(t, "unknown")

	if exitCode == 0 {
		t.Errorf("exit code = 0, want non-zero (알 수 없는 커맨드)")
	}
}
