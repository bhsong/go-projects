package main_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// moduleRoot는 stream-tool 모듈 루트 디렉토리를 반환한다.
// go test 실행 시 워킹 디렉토리는 cmd/ 이므로 한 단계 위로 올라간다.
func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	return filepath.Join(wd, "..")
}

// buildBinary는 stream-tool 바이너리를 임시 디렉토리에 빌드하고 경로를 반환한다.
// 모듈 루트에서 빌드해 모든 내부 패키지가 올바르게 해석되도록 한다.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "stream-tool")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd")
	cmd.Dir = moduleRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

type runResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runBinary는 bin을 dir 디렉토리에서 args와 함께 실행하고 결과를 반환한다.
func runBinary(t *testing.T, bin, dir string, args ...string) runResult {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected exec error: %v", err)
		}
	}
	return runResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: code,
	}
}

func TestE2E_파일입력(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	res := runBinary(t, bin, root, "testdata/sample.txt")

	if res.exitCode != 0 {
		t.Errorf("exit code = %d, want 0\nstderr: %s", res.exitCode, res.stderr)
	}
	// sample.txt = "hello world\nfoo bar baz\n" → Lines:2, Words:5, Bytes:24
	want := "줄: 2\n단어: 5\n바이트: 24\n"
	if res.stdout != want {
		t.Errorf("stdout = %q, want %q", res.stdout, want)
	}
}

func TestE2E_파일없음(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	res := runBinary(t, bin, root, "testdata/notexist.txt")

	if res.exitCode != 1 {
		t.Errorf("exit code = %d, want 1", res.exitCode)
	}
	wantStderr := "오류: open testdata/notexist.txt: no such file or directory\n"
	if res.stderr != wantStderr {
		t.Errorf("stderr = %q, want %q", res.stderr, wantStderr)
	}
}

func TestE2E_인자없음(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	res := runBinary(t, bin, root)

	if res.exitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.exitCode)
	}
	if !strings.Contains(res.stdout, "사용법:") {
		t.Errorf("stdout does not contain usage message:\n%s", res.stdout)
	}
}

func TestE2E_tee플래그(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)
	teeFile := filepath.Join(t.TempDir(), "tee.txt")

	res := runBinary(t, bin, root, "--tee", teeFile, "testdata/sample.txt")

	if res.exitCode != 0 {
		t.Errorf("exit code = %d, want 0\nstderr: %s", res.exitCode, res.stderr)
	}
	// tee.txt는 sample.txt 내용과 바이트 단위로 동일해야 한다.
	sampleBytes, err := os.ReadFile(filepath.Join(root, "testdata/sample.txt"))
	if err != nil {
		t.Fatalf("ReadFile sample.txt: %v", err)
	}
	teeBytes, err := os.ReadFile(teeFile)
	if err != nil {
		t.Fatalf("ReadFile tee.txt: %v", err)
	}
	if !bytes.Equal(teeBytes, sampleBytes) {
		t.Errorf("tee.txt (%q) != sample.txt (%q)", teeBytes, sampleBytes)
	}
}

func TestE2E_out플래그(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)
	outFile := filepath.Join(t.TempDir(), "out.txt")

	res := runBinary(t, bin, root, "--out", outFile, "testdata/sample.txt")

	if res.exitCode != 0 {
		t.Errorf("exit code = %d, want 0\nstderr: %s", res.exitCode, res.stderr)
	}
	// out.txt 내용은 stdout 내용과 동일해야 한다.
	outBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("ReadFile out.txt: %v", err)
	}
	if string(outBytes) != res.stdout {
		t.Errorf("out.txt = %q, stdout = %q: want equal", outBytes, res.stdout)
	}
	if res.stdout == "" {
		t.Error("stdout must not be empty")
	}
}
