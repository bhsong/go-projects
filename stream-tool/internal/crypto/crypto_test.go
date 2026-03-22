package crypto_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	mycrypto "github.com/bhsong/go-projects/stream-tool/internal/crypto"
)

// tempFile은 dir 아래 name 파일을 content로 생성하고 경로를 반환한다.
// t.TempDir()과 함께 사용하면 테스트 종료 시 자동 삭제된다.
func tempFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("tempFile: %v", err)
	}
	return path
}

// ---- Unit 테스트 ----

func TestHashFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	helloPath := tempFile(t, dir, "hello.txt", []byte("hello world"))
	emptyPath := tempFile(t, dir, "empty.txt", []byte{})

	sumHello := sha256.Sum256([]byte("hello world"))
	wantHello := hex.EncodeToString(sumHello[:])

	sumEmpty := sha256.Sum256([]byte{})
	wantEmpty := hex.EncodeToString(sumEmpty[:])

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "[정상] known content → expected SHA-256 hex",
			path: helloPath,
			want: wantHello,
		},
		{
			name: "[정상] empty file → SHA-256 of empty",
			path: emptyPath,
			want: wantEmpty,
		},
		{
			name:    "[엣지] file not found → error",
			path:    filepath.Join(dir, "nonexistent.txt"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := mycrypto.HashFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("hash = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVerifyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := tempFile(t, dir, "data.txt", []byte("hello world"))

	sum := sha256.Sum256([]byte("hello world"))
	correctHash := hex.EncodeToString(sum[:])
	// 처음 4자를 "aaaa"로 바꿔 틀린 해시를 만든다
	wrongHash := "aaaa" + correctHash[4:]

	tests := []struct {
		name    string
		path    string
		hash    string
		want    bool
		wantErr bool
	}{
		{
			name: "[정상] correct hash → returns true, nil",
			path: path,
			hash: correctHash,
			want: true,
		},
		{
			name: "[정상] wrong hash → returns false, nil",
			path: path,
			hash: wrongHash,
			want: false,
		},
		{
			name:    "[엣지] invalid hex string → returns false, error",
			path:    path,
			hash:    "not-valid-hex!!",
			wantErr: true,
		},
		{
			name:    "[엣지] file not found → returns false, error",
			path:    filepath.Join(dir, "nonexistent.txt"),
			hash:    correctHash,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := mycrypto.VerifyFile(tt.path, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ok = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateHMAC(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file1 := tempFile(t, dir, "file1.txt", []byte("content A"))
	file2 := tempFile(t, dir, "file2.txt", []byte("content B"))

	// [정상] 케이스의 기준값: file1 + "secret" 조합으로 생성한 MAC
	refMAC, err := mycrypto.GenerateHMAC(file1, "secret")
	if err != nil {
		t.Fatalf("setup GenerateHMAC: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		key       string
		wantEqual bool // refMAC와 같아야 하는 경우 (결정적)
		wantDiff  bool // refMAC와 달라야 하는 경우
		wantErr   bool
	}{
		{
			name:      "[정상] same file + same key → same MAC (deterministic)",
			path:      file1,
			key:       "secret",
			wantEqual: true,
		},
		{
			name:     "[정상] same file + diff key → different MAC",
			path:     file1,
			key:      "other-secret",
			wantDiff: true,
		},
		{
			name:     "[정상] diff file + same key → different MAC",
			path:     file2,
			key:      "secret",
			wantDiff: true,
		},
		{
			name:    "[엣지] file not found → error",
			path:    filepath.Join(dir, "nonexistent.txt"),
			key:     "secret",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := mycrypto.GenerateHMAC(tt.path, tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.wantEqual && got != refMAC {
				t.Errorf("MAC = %q, want %q (should be equal to refMAC)", got, refMAC)
			}
			if tt.wantDiff && got == refMAC {
				t.Errorf("MAC = %q (should differ from refMAC %q)", got, refMAC)
			}
		})
	}
}

func TestVerifyHMAC(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := tempFile(t, dir, "data.txt", []byte("test content"))

	correctMAC, err := mycrypto.GenerateHMAC(path, "secret")
	if err != nil {
		t.Fatalf("setup GenerateHMAC (correct key): %v", err)
	}
	wrongKeyMAC, err := mycrypto.GenerateHMAC(path, "wrong-key")
	if err != nil {
		t.Fatalf("setup GenerateHMAC (wrong key): %v", err)
	}

	tests := []struct {
		name    string
		path    string
		key     string
		mac     string
		want    bool
		wantErr bool
	}{
		{
			name: "[정상] correct key + correct mac → true, nil",
			path: path,
			key:  "secret",
			mac:  correctMAC,
			want: true,
		},
		{
			name: "[정상] correct key + wrong mac → false, nil",
			path: path,
			key:  "secret",
			mac:  wrongKeyMAC,
			want: false,
		},
		{
			name: "[정상] wrong key + any mac → false, nil",
			path: path,
			key:  "wrong-key",
			mac:  correctMAC,
			want: false,
		},
		{
			name:    "[엣지] invalid hex mac string → false, error",
			path:    path,
			key:     "secret",
			mac:     "not-valid-hex!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := mycrypto.VerifyHMAC(tt.path, tt.key, tt.mac)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ok = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	original := []byte("this is my secret content")
	srcPath := tempFile(t, dir, "plain.txt", original)
	// 44바이트 미만(43B) → ErrFileTooShort 발생용
	tooShortPath := tempFile(t, dir, "short.bin", make([]byte, 43))

	tests := []struct {
		name            string
		srcPath         string
		password        string
		decPassword     string
		setupEncrypt    bool  // 먼저 EncryptFile을 실행해야 하는 경우
		skipDecrypt     bool  // DecryptFile 단계를 건너뛸 경우
		wantEncryptErr  bool  // EncryptFile이 에러를 반환해야 하는 경우
		wantDecryptErr  error // errors.Is로 확인할 sentinel 에러
	}{
		{
			name:         "[정상] encrypt then decrypt with same password → plaintext matches",
			srcPath:      srcPath,
			password:     "correct-pass",
			decPassword:  "correct-pass",
			setupEncrypt: true,
		},
		{
			name:         "[정상] encrypted file differs from original (not plaintext)",
			srcPath:      srcPath,
			password:     "some-pass",
			setupEncrypt: true,
			skipDecrypt:  true,
		},
		{
			name:           "[엣지] decrypt with wrong password → ErrDecryptFailed",
			srcPath:        srcPath,
			password:       "correct-pass",
			decPassword:    "wrong-pass",
			setupEncrypt:   true,
			wantDecryptErr: mycrypto.ErrDecryptFailed,
		},
		{
			name:           "[엣지] decrypt file too short → ErrFileTooShort",
			srcPath:        tooShortPath,
			decPassword:    "any-pass",
			wantDecryptErr: mycrypto.ErrFileTooShort,
		},
		{
			name:           "[엣지] src file not found → error on EncryptFile",
			srcPath:        filepath.Join(dir, "nonexistent.txt"),
			password:       "pass",
			setupEncrypt:   true,
			wantEncryptErr: true,
			skipDecrypt:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			encPath := filepath.Join(t.TempDir(), "encrypted.bin")
			decPath := filepath.Join(t.TempDir(), "decrypted.txt")

			// EncryptFile 단계
			if tt.setupEncrypt || tt.wantEncryptErr {
				encErr := mycrypto.EncryptFile(tt.srcPath, encPath, tt.password)
				if tt.wantEncryptErr {
					if encErr == nil {
						t.Fatal("expected EncryptFile to return error, got nil")
					}
					return
				}
				if encErr != nil {
					t.Fatalf("EncryptFile: %v", encErr)
				}
				// 암호화 파일은 원본과 달라야 한다
				enc, err := os.ReadFile(encPath)
				if err != nil {
					t.Fatalf("read encrypted file: %v", err)
				}
				if bytes.Equal(enc, original) {
					t.Error("encrypted file must differ from original plaintext")
				}
			}

			if tt.skipDecrypt {
				return
			}

			// DecryptFile 단계: 사전 암호화한 경우 encPath, 아닌 경우 srcPath를 사용
			decSrc := encPath
			if !tt.setupEncrypt {
				decSrc = tt.srcPath
			}

			decErr := mycrypto.DecryptFile(decSrc, decPath, tt.decPassword)
			if tt.wantDecryptErr != nil {
				if !errors.Is(decErr, tt.wantDecryptErr) {
					t.Errorf("DecryptFile err = %v, want errors.Is(%v)", decErr, tt.wantDecryptErr)
				}
				return
			}
			if decErr != nil {
				t.Fatalf("DecryptFile: %v", decErr)
			}

			// 복호화 결과가 원본과 일치해야 한다
			got, err := os.ReadFile(decPath)
			if err != nil {
				t.Fatalf("read decrypted file: %v", err)
			}
			if !bytes.Equal(got, original) {
				t.Errorf("decrypted = %q, want %q", got, original)
			}
		})
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	srcPath := tempFile(t, dir, "plain.txt", []byte("same content every time"))

	tests := []struct {
		name string
	}{
		{name: "[정상] same input encrypted twice → different ciphertext (salt+nonce 매번 다름)"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			enc1 := filepath.Join(t.TempDir(), "enc1.bin")
			enc2 := filepath.Join(t.TempDir(), "enc2.bin")

			if err := mycrypto.EncryptFile(srcPath, enc1, "password"); err != nil {
				t.Fatalf("EncryptFile 1: %v", err)
			}
			if err := mycrypto.EncryptFile(srcPath, enc2, "password"); err != nil {
				t.Fatalf("EncryptFile 2: %v", err)
			}

			b1, err := os.ReadFile(enc1)
			if err != nil {
				t.Fatalf("read enc1: %v", err)
			}
			b2, err := os.ReadFile(enc2)
			if err != nil {
				t.Fatalf("read enc2: %v", err)
			}
			if bytes.Equal(b1, b2) {
				t.Error("two encryptions of the same file must produce different ciphertext (salt/nonce not reused)")
			}
		})
	}
}

// ---- Feature 테스트 ----

func TestFeature_HashAndVerifyPipeline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := tempFile(t, dir, "data.txt", []byte("pipeline test data"))

	tests := []struct {
		name string
	}{
		{name: "[정상] HashFile 결과를 VerifyFile에 바로 사용 → true"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hash, err := mycrypto.HashFile(path)
			if err != nil {
				t.Fatalf("HashFile: %v", err)
			}
			ok, err := mycrypto.VerifyFile(path, hash)
			if err != nil {
				t.Fatalf("VerifyFile: %v", err)
			}
			if !ok {
				t.Error("VerifyFile with freshly computed hash must return true")
			}
		})
	}
}

func TestFeature_EncryptDecryptLargeFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// 1MB 랜덤 데이터 생성 — 실제 파일 I/O 파이프라인 전체 경로를 검증한다
	data := make([]byte, 1024*1024)
	if _, err := rand.Read(data); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	srcPath := tempFile(t, dir, "large.bin", data)

	tests := []struct {
		name string
	}{
		{name: "[정상] 1MB 데이터 암호화/복호화 → 내용 일치"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			encPath := filepath.Join(t.TempDir(), "large.enc")
			decPath := filepath.Join(t.TempDir(), "large.dec")

			if err := mycrypto.EncryptFile(srcPath, encPath, "large-file-password"); err != nil {
				t.Fatalf("EncryptFile: %v", err)
			}
			if err := mycrypto.DecryptFile(encPath, decPath, "large-file-password"); err != nil {
				t.Fatalf("DecryptFile: %v", err)
			}

			got, err := os.ReadFile(decPath)
			if err != nil {
				t.Fatalf("read decrypted: %v", err)
			}
			if !bytes.Equal(got, data) {
				t.Error("decrypted large file must match original")
			}
		})
	}
}
