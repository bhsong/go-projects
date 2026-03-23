package main_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

	res := runBinary(t, bin, root, "count", "testdata/sample.txt")

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

	res := runBinary(t, bin, root, "count", "testdata/notexist.txt")

	if res.exitCode != 1 {
		t.Errorf("exit code = %d, want 1", res.exitCode)
	}
	if !strings.Contains(res.stderr, "no such file or directory") {
		t.Errorf("stderr = %q, want to contain %q", res.stderr, "no such file or directory")
	}
}

func TestE2E_stdin입력(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	cmd := exec.Command(bin, "count")
	cmd.Dir = root
	cmd.Stdin = strings.NewReader("one two\nthree\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v\nstderr: %s", err, stderr.String())
	}
	// "one two\n"(8) + "three\n"(6) → Lines:2, Words:3, Bytes:14
	want := "줄: 2\n단어: 3\n바이트: 14\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestE2E_tee플래그(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)
	teeFile := filepath.Join(t.TempDir(), "tee.txt")

	res := runBinary(t, bin, root, "count", "--tee", teeFile, "testdata/sample.txt")

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

	res := runBinary(t, bin, root, "count", "--out", outFile, "testdata/sample.txt")

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

// ---- Crypto e2e 테스트 ----

func TestE2E_hash(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	// testdata/sample.txt의 SHA-256 기준값 (표준 라이브러리로 계산)
	sampleData, err := os.ReadFile(filepath.Join(root, "testdata/sample.txt"))
	if err != nil {
		t.Fatalf("read sample.txt: %v", err)
	}
	sum := sha256.Sum256(sampleData)
	wantHash := hex.EncodeToString(sum[:])

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantHash string // 비어 있지 않으면 stdout과 비교
		wantHex  bool   // stdout이 64자 hex여야 하는 경우
	}{
		{
			name:     "[정상] hash <file> → 64자 hex 출력, exit 0",
			args:     []string{"hash", "testdata/sample.txt"},
			wantCode: 0,
			wantHash: wantHash,
			wantHex:  true,
		},
		{
			name:     "[엣지] 파일 없음 → exit 1",
			args:     []string{"hash", "testdata/nonexistent.txt"},
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runBinary(t, bin, root, tt.args...)
			if res.exitCode != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstderr: %s", res.exitCode, tt.wantCode, res.stderr)
			}
			if tt.wantHex {
				got := strings.TrimSpace(res.stdout)
				if len(got) != 64 {
					t.Errorf("stdout = %q, want 64-char hex", got)
				}
				if _, decErr := hex.DecodeString(got); decErr != nil {
					t.Errorf("stdout is not valid hex: %q", got)
				}
			}
			if tt.wantHash != "" {
				got := strings.TrimSpace(res.stdout)
				if got != tt.wantHash {
					t.Errorf("hash = %q, want %q", got, tt.wantHash)
				}
			}
		})
	}
}

func TestE2E_verify(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	sampleData, err := os.ReadFile(filepath.Join(root, "testdata/sample.txt"))
	if err != nil {
		t.Fatalf("read sample.txt: %v", err)
	}
	sum := sha256.Sum256(sampleData)
	correctHash := hex.EncodeToString(sum[:])
	wrongHash := "aaaa" + correctHash[4:]

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string // stdout에 포함되어야 할 문자열 (비어 있으면 검사 생략)
		wantErr  string // stderr에 포함되어야 할 문자열 (비어 있으면 검사 생략)
	}{
		{
			name:     "[정상] 올바른 hash → OK: hash matches, exit 0",
			args:     []string{"verify", "testdata/sample.txt", correctHash},
			wantCode: 0,
			wantOut:  "OK: hash matches",
		},
		{
			name:     "[정상] 틀린 hash → FAIL: hash mismatch, exit 1",
			args:     []string{"verify", "testdata/sample.txt", wrongHash},
			wantCode: 1,
			wantErr:  "FAIL: hash mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runBinary(t, bin, root, tt.args...)
			if res.exitCode != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", res.exitCode, tt.wantCode, res.stdout, res.stderr)
			}
			if tt.wantOut != "" && !strings.Contains(res.stdout, tt.wantOut) {
				t.Errorf("stdout = %q, want to contain %q", res.stdout, tt.wantOut)
			}
			if tt.wantErr != "" && !strings.Contains(res.stderr, tt.wantErr) {
				t.Errorf("stderr = %q, want to contain %q", res.stderr, tt.wantErr)
			}
		})
	}
}

func TestE2E_hmac(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantHex  bool // stdout이 64자 hex여야 하는 경우
	}{
		{
			name:     "[정상] --key 있음 → 64자 hex 출력, exit 0",
			args:     []string{"hmac", "--key", "testkey", "testdata/sample.txt"},
			wantCode: 0,
			wantHex:  true,
		},
		{
			name:     "[엣지] --key 없음 → exit 1",
			args:     []string{"hmac", "testdata/sample.txt"},
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runBinary(t, bin, root, tt.args...)
			if res.exitCode != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstderr: %s", res.exitCode, tt.wantCode, res.stderr)
			}
			if tt.wantHex {
				got := strings.TrimSpace(res.stdout)
				if len(got) != 64 {
					t.Errorf("stdout = %q, want 64-char hex", got)
				}
				if _, decErr := hex.DecodeString(got); decErr != nil {
					t.Errorf("stdout is not valid hex: %q", got)
				}
			}
		})
	}
}

func TestE2E_hmac_verify(t *testing.T) {
	bin := buildBinary(t)
	root := moduleRoot(t)

	// 올바른 MAC을 얻기 위해 먼저 hmac 서브커맨드를 실행한다.
	// 구현 전에는 이 단계에서 실패하므로 테스트는 Red 상태가 된다.
	hmacRes := runBinary(t, bin, root, "hmac", "--key", "testkey", "testdata/sample.txt")
	if hmacRes.exitCode != 0 {
		t.Fatalf("setup: hmac subcommand returned exit %d (not yet implemented?)\nstderr: %s",
			hmacRes.exitCode, hmacRes.stderr)
	}
	correctMAC := strings.TrimSpace(hmacRes.stdout)
	wrongMAC := "aaaa" + correctMAC[4:]

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "[정상] 올바른 mac → OK: hmac matches, exit 0",
			args:     []string{"hmac-verify", "--key", "testkey", "testdata/sample.txt", correctMAC},
			wantCode: 0,
			wantOut:  "OK: hmac matches",
		},
		{
			name:     "[정상] 틀린 mac → FAIL: hmac mismatch, exit 1",
			args:     []string{"hmac-verify", "--key", "testkey", "testdata/sample.txt", wrongMAC},
			wantCode: 1,
			wantErr:  "FAIL: hmac mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runBinary(t, bin, root, tt.args...)
			if res.exitCode != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s", res.exitCode, tt.wantCode, res.stdout, res.stderr)
			}
			if tt.wantOut != "" && !strings.Contains(res.stdout, tt.wantOut) {
				t.Errorf("stdout = %q, want to contain %q", res.stdout, tt.wantOut)
			}
			if tt.wantErr != "" && !strings.Contains(res.stderr, tt.wantErr) {
				t.Errorf("stderr = %q, want to contain %q", res.stderr, tt.wantErr)
			}
		})
	}
}

func TestE2E_encrypt_decrypt_roundtrip(t *testing.T) {
	bin := buildBinary(t)

	dir := t.TempDir()
	content := []byte("secret message for e2e roundtrip test")
	srcPath := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(srcPath, content, 0600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	tests := []struct {
		name        string
		password    string
		decPassword string
		wantEncCode int
		wantDecCode int
		wantMatch   bool // 복호화된 내용이 원본과 일치해야 하는 경우
	}{
		{
			name:        "[정상] encrypt → decrypt → 원본과 일치, exit 0",
			password:    "my-secret-password",
			decPassword: "my-secret-password",
			wantEncCode: 0,
			wantDecCode: 0,
			wantMatch:   true,
		},
		{
			name:        "[엣지] decrypt with wrong password → exit 1",
			password:    "my-secret-password",
			decPassword: "wrong-password",
			wantEncCode: 0,
			wantDecCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encPath := filepath.Join(t.TempDir(), "encrypted.bin")
			decPath := filepath.Join(t.TempDir(), "decrypted.txt")

			encRes := runBinary(t, bin, dir,
				"encrypt", "--pass", tt.password, "--out", encPath, srcPath)
			if encRes.exitCode != tt.wantEncCode {
				t.Fatalf("encrypt exit code = %d, want %d\nstderr: %s",
					encRes.exitCode, tt.wantEncCode, encRes.stderr)
			}

			decRes := runBinary(t, bin, dir,
				"decrypt", "--pass", tt.decPassword, "--out", decPath, encPath)
			if decRes.exitCode != tt.wantDecCode {
				t.Errorf("decrypt exit code = %d, want %d\nstderr: %s",
					decRes.exitCode, tt.wantDecCode, decRes.stderr)
			}

			if tt.wantMatch {
				got, err := os.ReadFile(decPath)
				if err != nil {
					t.Fatalf("read decrypted: %v", err)
				}
				if !bytes.Equal(got, content) {
					t.Errorf("decrypted content = %q, want %q", got, content)
				}
			}
		})
	}
}
