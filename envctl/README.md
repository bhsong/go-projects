# envctl

`.env` 파일을 파싱, 검증, 병합하고 환경변수를 주입하여 명령어를 실행하는 CLI 도구.

```
$ envctl parse .env
DB_HOST              = localhost
DB_PORT              = 5432
URL                  = localhost:5432/db

$ envctl exec .env -- go run main.go
```

---

## 설치

```bash
go install github.com/bhsong/go-projects/envctl/cmd@latest
```

또는 소스에서 빌드:

```bash
git clone https://github.com/bhsong/go-projects/envctl
cd envctl
go build -o envctl ./cmd
```

---

## 사용법

```
사용법: envctl <명령어> [옵션]

명령어:
  parse  <파일>                              .env 파일 파싱 결과 출력 (변수 치환 포함)
  check  <파일> [파일...]                    중복 키 검출
  merge  <기준파일> <오버라이드파일> [...]    우선순위 병합 결과 출력
  exec   <파일> [파일...] -- <명령어> [인자...]
                                             환경변수 주입 후 명령 실행
```

### `parse` — 파싱 및 변수 치환 결과 출력

`.env` 파일을 읽어 `$VAR` / `${VAR}` 치환까지 완료된 결과를 정렬된 키 순서로 출력합니다.

```bash
envctl parse .env
envctl parse config/production.env
```

**예시 `.env`:**
```env
DB_HOST=localhost
DB_PORT=5432
DATABASE_URL=$DB_HOST:$DB_PORT/mydb
```

**출력:**
```
DATABASE_URL         = localhost:5432/mydb
DB_HOST              = localhost
DB_PORT              = 5432
```

---

### `check` — 중복 키 검출

하나 이상의 `.env` 파일에서 동일한 키가 여러 번 정의된 경우를 탐지합니다.

```bash
envctl check .env
envctl check .env .env.local .env.production
```

**출력 예시 (중복 없음):**
```
[.env] 중복 없음 ✓
```

**출력 예시 (중복 있음):**
```
[.env]
  ⚠  DB_HOST: 1, 7번째 줄에서 중복 정의됨
```

---

### `merge` — 우선순위 병합

여러 `.env` 파일을 왼쪽→오른쪽 우선순위로 병합합니다. 오른쪽 파일의 값이 왼쪽을 덮어씁니다. 결과는 `KEY=VALUE` 형태로 출력됩니다 (shell `source` 호환).

```bash
# .env.local이 .env를 오버라이드
envctl merge .env .env.local

# 3단계 우선순위: base < staging < local
envctl merge .env .env.staging .env.local
```

**예시:**
```bash
# .env      : DB_HOST=prod-db, PORT=5432
# .env.local: DB_HOST=localhost

envctl merge .env .env.local
# DB_HOST=localhost   ← .env.local 우선
# PORT=5432           ← .env에서 가져옴
```

---

### `exec` — 환경변수 주입 후 명령 실행

`.env` 파일(들)을 로드하고 환경변수를 주입한 채 명령어를 실행합니다. `--`로 파일 목록과 실행 명령어를 구분합니다.

```bash
# 단일 파일
envctl exec .env -- go run main.go

# 여러 파일 병합 후 실행 (오른쪽 우선)
envctl exec .env .env.local -- go test ./...

# 환경변수 확인
envctl exec .env -- env | grep DB_
```

> 기존 프로세스 환경변수는 유지되며, `.env` 키와 충돌할 경우 기존 값을 보존합니다.

---

## `.env` 파일 형식

```env
# 주석은 # 으로 시작
DB_HOST=localhost
DB_PORT=5432

# 빈 줄 허용
EMPTY_VALUE=

# 변수 치환: $VAR 또는 ${VAR}
DATABASE_URL=$DB_HOST:$DB_PORT/mydb
BACKUP_URL=${DB_HOST}:${DB_PORT}/backup
```

| 규칙 | 설명 |
|---|---|
| `KEY=VALUE` | 기본 형식. `=` 앞뒤 공백은 제거됨 |
| `# 주석` | `#`으로 시작하는 줄은 무시 |
| 빈 줄 | 무시 |
| `EMPTY=` | 빈 값 허용 |
| `$VAR`, `${VAR}` | 같은 파일 내 다른 키를 참조하여 치환 |
| 존재하지 않는 키 참조 | 빈 문자열로 치환 (bash 기본 동작과 동일) |

---

## 프로젝트 구조

```
envctl/
├── go.mod
├── cmd/
│   ├── main.go          # CLI 진입점: 서브커맨드 분기, 파일 I/O, 출력
│   └── e2e_test.go      # E2E 테스트: 바이너리 빌드 후 exec.Command로 검증
└── internal/
    └── env/
        ├── errors.go    # EnvParseError — errors.As 호환 커스텀 에러
        ├── parser.go    # Entry, Parse, parseLine — io.Reader 기반 파서
        ├── mapper.go    # ToMap, MapToSlice — []Entry ↔ map 변환
        ├── expander.go  # Expand — $VAR / ${VAR} 치환
        ├── merger.go    # Merge, Apply — 병합 및 os.Setenv 적용
        └── env_test.go  # Unit + Feature 테스트
```

**설계 원칙:**

- `internal/env` — 파일시스템과 무관한 순수 도메인 로직. `io.Reader`를 받아 테스트 주입 가능.
- `cmd/main.go` — `os.Open`, `os.Args`, 출력만 담당. 도메인 로직을 직접 구현하지 않음.

---

## 개발

```bash
# 테스트 실행
go test ./...

# 경쟁 조건 검사
go test -race ./...

# 포맷 / 정적 분석
go fmt ./...
go vet ./...
```
