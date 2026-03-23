# go-projects

Go 언어를 시스템/네트워크/인프라 관점에서 학습하는 사이드 프로젝트 모음.

---

## 프로젝트 목록

### [todo-cli](./todo-cli)

터미널에서 동작하는 할 일 관리 CLI 도구.

- `add` / `list` / `done` / `delete` 서브커맨드 지원
- JSON 파일로 데이터 영속 저장 (Atomic write 패턴 적용)
- `flag.NewFlagSet` 기반 서브커맨드별 플래그 독립 관리
- 표준 라이브러리만 사용 (외부 의존성 없음)
- Unit / Feature / E2E 3단계 테스트 구성

**학습 핵심:** `struct` · 파일 I/O · error 처리 · `os.Args` · JSON 직렬화 · `flag` 패키지 · 서브커맨드 패턴

---

### [expr-calc](./expr-calc)

수식 문자열을 파싱하고 계산하는 CLI 도구.

- 사칙연산(`+` `-` `*` `/`)과 괄호 지원, 연산자 우선순위 자동 처리
- 재귀 하강 파서(Recursive Descent Parser) 구현
- `defer` / `panic` / `recover` 패턴으로 파서 내부 에러를 공개 API 경계에서 `error`로 변환
- `[]rune` 기반 Lexer로 유니코드 안전 처리
- 표준 라이브러리만 사용 (외부 의존성 없음)

**학습 핵심:** `defer` · `panic` · `recover` · 재귀 하강 파싱 · Lexer/Parser 설계

---

### [envctl](./envctl)

`.env` 파일을 파싱, 검증, 병합하고 환경변수를 주입하여 명령어를 실행하는 CLI 도구.

- `parse` / `check` / `merge` / `exec` 서브커맨드 지원
- `$VAR` / `${VAR}` 변수 치환 (파일 내 자기 참조)
- 중복 키 검출 및 다중 파일 우선순위 병합
- `io.Reader` 기반 파서로 도메인 로직과 파일 I/O 완전 분리
- 커스텀 에러 타입 (`errors.As` 호환)
- 표준 라이브러리만 사용 (외부 의존성 없음)

**학습 핵심:** `io.Reader` 인터페이스 · 커스텀 에러 타입 · `bufio.Scanner` · `os.Setenv` / `exec.Command` · 도메인 계층 분리 · `syscall.Exec`

---

### [stream-tool](./stream-tool)

파일 또는 stdin을 읽어 통계(줄/단어/바이트)를 집계하고, SHA-256 해시·HMAC·AES-256-GCM 암호화를 수행하는 CLI 도구.

**Stream 서브커맨드**

- `count [file]` — 줄/단어/바이트 집계 (인자 없으면 stdin)
- `--tee <file>` — 입력 내용을 파일에 동시 저장하며 집계
- `--out <file>` — 집계 결과를 파일과 stdout에 동시 출력

**Crypto 서브커맨드**

- `hash <file>` — SHA-256 해시 출력
- `verify <file> <hash>` — SHA-256 해시 검증
- `hmac --key <secret> <file>` — HMAC-SHA256 생성
- `hmac-verify --key <secret> <file> <mac>` — HMAC-SHA256 검증
- `encrypt --pass <pw> --out <dst> <src>` — AES-256-GCM 파일 암호화 (scrypt KDF)
- `decrypt --pass <pw> --out <dst> <src>` — AES-256-GCM 파일 복호화

**학습 핵심:** `io.Reader/Writer` 파이프라인 · `io.TeeReader` · `io.MultiWriter` · `crypto/sha256` · `crypto/hmac` · `crypto/aes` (GCM) · `scrypt` 키 파생 · `subtle.ConstantTimeCompare` · `crypto/rand`
