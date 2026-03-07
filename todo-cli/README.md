# todo-cli

터미널에서 동작하는 할 일 관리 CLI 도구. Go 표준 라이브러리만 사용하며, JSON 파일로 데이터를 영속적으로 저장합니다.

## 요구사항

- Go 1.21+

## 설치

```bash
git clone https://github.com/bhsong/go-projects/todo-cli
cd todo-cli
go build -o todo ./cmd/
```

빌드된 바이너리를 PATH에 추가하면 어디서든 실행할 수 있습니다:

```bash
mv todo /usr/local/bin/
```

## 사용법

```
todo <command> [arguments]
```

| 커맨드 | 설명 | 예시 |
|---|---|---|
| `add <title>` | 새 할 일 추가 | `todo add "PR 리뷰하기"` |
| `list` | 전체 목록 조회 | `todo list` |
| `done <id>` | 할 일 완료 처리 | `todo done 1` |
| `delete <id>` | 할 일 삭제 | `todo delete 2` |

### 실행 예시

```bash
$ todo add "PR 리뷰하기"
추가됨: PR 리뷰하기

$ todo add "배포 스크립트 수정"
추가됨: 배포 스크립트 수정

$ todo list
⬜ [1] PR 리뷰하기
⬜ [2] 배포 스크립트 수정

$ todo done 1
완료 처리: ID 1

$ todo list
✅ [1] PR 리뷰하기
⬜ [2] 배포 스크립트 수정

$ todo delete 2
삭제됨: ID 2
```

데이터는 실행 디렉터리의 `tasks.json`에 저장됩니다. 프로그램을 재시작해도 목록이 유지됩니다.

## 프로젝트 구조

```
todo-cli/
├── cmd/
│   ├── main.go           # 진입점 — 커맨드 파싱, 헬퍼 함수
│   └── main_test.go      # 헬퍼 함수 단위 테스트
├── internal/
│   ├── task/
│   │   ├── task.go       # Task 구조체 + 비즈니스 로직 (Add/Complete/Delete)
│   │   └── task_test.go  # 단위 테스트
│   ├── storage/
│   │   ├── storage.go    # JSON 파일 읽기/쓰기 (Load/Save)
│   │   └── storage_test.go
│   └── feature_test.go   # 기능 통합 테스트
└── testdata/
    └── e2e_test.go       # E2E 테스트 (바이너리 빌드 후 실행)
```

`internal/` 패키지는 외부 모듈에서 import할 수 없도록 컴파일러가 강제합니다. `task`와 `storage`는 서로를 모르며, `cmd/main.go`가 두 패키지를 조율합니다. 이 구조 덕분에 나중에 JSON 대신 DB로 교체할 때 `storage` 패키지만 수정하면 됩니다.

## 데이터 포맷

```json
[
  {
    "id": 1,
    "title": "PR 리뷰하기",
    "done": true,
    "created_at": "2026-03-07T09:00:00+09:00"
  }
]
```

## 테스트

```bash
# 단위 + 기능 테스트
go test ./...

# E2E 테스트 (바이너리 빌드 포함)
go test ./testdata/

# 레이스 컨디션 검사
go test -race ./...

# 상세 출력
go test -v ./...
```

테스트는 3개 레벨로 구성됩니다:

- **Unit** — 함수 하나의 입출력 검증 (파일 I/O 없음)
- **Feature** — `task` + `storage` 패키지 협력 검증 (`t.TempDir()`로 파일 격리)
- **E2E** — 실제 바이너리를 빌드해서 CLI 명령어 실행 및 stdout/stderr/exit code 검증

## 에러 처리

존재하지 않는 ID를 조작하면 에러 메시지를 출력하고 exit code 1로 종료합니다:

```bash
$ todo done 999
오류: task.Complete: ID 999 없음

$ echo $?
1
```

숫자가 아닌 ID를 입력해도 동일하게 처리됩니다:

```bash
$ todo done abc
숫자가 아닌 ID: strconv.Atoi: parsing "abc": invalid syntax
```
