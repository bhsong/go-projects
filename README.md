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
