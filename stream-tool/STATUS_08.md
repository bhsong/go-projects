# STATUS 08 — stream-tool crypto 기능 추가

## 스테이지 정보

| 항목 | 내용 |
|------|------|
| 스테이지 | 08 |
| 유형 | 기능 추가 (기존 stream-tool 위에 crypto 서브커맨드 추가) |
| 브랜치 | feat/08-crypto |
| 작업 위치 | stream-tool/ |
| 완료일 | 2026-03-23 |

---

## 학습 목표

- `crypto/sha256` + `io.Writer` 조합으로 스트리밍 해시 계산 이해
- `crypto/hmac`으로 메시지 인증 코드(MAC) 생성/검증
- `crypto/aes` + GCM 모드로 인증된 암호화(AEAD) 구현
- `golang.org/x/crypto/scrypt`로 비밀번호 기반 키 파생(KDF)
- `crypto/rand`로 nonce·salt 안전하게 생성
- `subtle.ConstantTimeCompare`로 타이밍 공격 방지

---

## 구현 내용

### 신규 파일

| 파일 | 설명 |
|------|------|
| `internal/crypto/hash.go` | SHA-256 해시 계산(`HashFile`) 및 검증(`VerifyFile`) |
| `internal/crypto/hmac.go` | HMAC-SHA256 생성(`GenerateHMAC`) 및 검증(`VerifyHMAC`) |
| `internal/crypto/aes.go` | AES-256-GCM 파일 암호화(`EncryptFile`) / 복호화(`DecryptFile`) |
| `internal/crypto/errors.go` | sentinel 에러 변수 정의 |
| `internal/crypto/crypto_test.go` | unit + feature 테스트 (테이블 드리븐) |

### 변경 파일

| 파일 | 설명 |
|------|------|
| `cmd/main.go` | `hash` / `verify` / `hmac` / `hmac-verify` / `encrypt` / `decrypt` 서브커맨드 추가 |
| `cmd/e2e_test.go` | 위 6개 서브커맨드 e2e 테스트 추가 |

---

## 주요 설계 결정

### AES-256-GCM 파일 포맷

```
[ salt (32 bytes) ][ nonce (12 bytes) ][ ciphertext + GCM tag ]
```

- salt: scrypt 키 파생용 랜덤 값 (매 암호화마다 새로 생성)
- nonce: GCM 모드 초기화 벡터 (매 암호화마다 새로 생성)
- GCM tag: 복호화 시 무결성 자동 검증

### scrypt 파라미터

```go
scrypt.Key([]byte(password), salt, N=32768, r=8, p=1, keyLen=32)
```

- N=32768(2^15): 메모리·CPU 비용 균형 (OWASP 권장 최솟값)
- keyLen=32: AES-256용 32바이트 키

### 타이밍 공격 방지

- `subtle.ConstantTimeCompare` — SHA-256 검증
- `hmac.Equal` — HMAC 검증 (내부적으로 constant-time 비교)

---

## 테스트 결과

```
$ go test ./...
ok  github.com/bhsong/go-projects/stream-tool/cmd
ok  github.com/bhsong/go-projects/stream-tool/internal/crypto
ok  github.com/bhsong/go-projects/stream-tool/internal/stream

$ go test -race ./...
ok  github.com/bhsong/go-projects/stream-tool/cmd
ok  github.com/bhsong/go-projects/stream-tool/internal/crypto
ok  github.com/bhsong/go-projects/stream-tool/internal/stream
```

---

## CLI 사용 예시

```bash
# SHA-256 해시
./stream-tool hash testdata/sample.txt

# 해시 검증
./stream-tool verify testdata/sample.txt <hash>

# HMAC 생성
./stream-tool hmac --key mysecret testdata/sample.txt

# HMAC 검증
./stream-tool hmac-verify --key mysecret testdata/sample.txt <mac>

# AES-256-GCM 암호화
./stream-tool encrypt --pass mypassword --out testdata/sample.enc testdata/sample.txt

# AES-256-GCM 복호화
./stream-tool decrypt --pass mypassword --out testdata/sample.dec testdata/sample.enc
```
