# expr-calc

수식 문자열을 파싱하고 계산하는 CLI 도구.

사칙연산(`+` `-` `*` `/`)과 괄호를 지원하며, 재귀 하강 파서(Recursive Descent Parser)와 Go의 `defer`/`panic`/`recover` 패턴을 사용하여 구현되었다.

---

## 설치 및 빌드

```bash
git clone https://github.com/bhsong/go-projects
cd go-projects/expr-calc
go build -o expr-calc ./cmd/
```

Go 1.21 이상 필요. 외부 의존성 없음.

---

## 사용법

```bash
./expr-calc "<표현식>"
```

인자를 여러 개로 분리해도 동작한다.

```bash
./expr-calc 3 + 4 * 2
```

### 예시

```
$ ./expr-calc "3 + 4 * 2"
11

$ ./expr-calc "(10 - 2) / 4"
2

$ ./expr-calc "1.5 + 2.5"
4

$ ./expr-calc "((2 + 3)) * 4"
20

$ ./expr-calc "(10 - 2) / 4 + 1 * 3"
5
```

### 에러 처리

```
$ ./expr-calc "1 / 0"
오류: calc.Parse: 0으로 나눌 수 없습니다.

$ ./expr-calc "(3 + 4"
오류: calc.Parser: 닫는 괄호가 없습니다.

$ ./expr-calc "3 + * 2"
오류: calc.Parser: 예상치 못한 토큰: "*"

$ ./expr-calc "3 @ 4"
오류: calc.Lexer: 알수 없는 문자: '@'
```

인자 없이 실행하면 사용법을 출력하고 종료 코드 0으로 종료된다.
에러 발생 시 stderr에 출력하고 종료 코드 1로 종료된다.

---

## 지원 문법

```
expr   = term   { ("+" | "-") term }
term   = factor { ("*" | "/") factor }
factor = NUMBER | "(" expr ")"
NUMBER = [0-9]+ ("." [0-9]+)?
```

연산자 우선순위는 문법 구조로 표현된다. `*`/`/`는 `+`/`-`보다 먼저 처리되며, 괄호로 우선순위를 변경할 수 있다.

---

## 프로젝트 구조

```
expr-calc/
├── go.mod
├── cmd/
│   └── main.go              # CLI 진입점
└── internal/
    └── calc/
        ├── lexer.go         # Lexer — 문자열을 Token 스트림으로 변환
        ├── parser.go        # Parser + Eval — 재귀 하강 파싱 및 계산
        └── calc_test.go     # 단위 테스트
```

---

## 아키텍처

### panic/recover 경계

파서 내부는 오류 발생 시 `parsePanic` 타입으로 panic을 던진다. `Eval()`이 공개 API 경계선 역할을 하며, `defer`/`recover`로 `parsePanic`을 `error`로 변환하여 외부에 노출한다.

```
[문자열 입력]
      │
      ▼
Eval(expr string) (float64, error)   ← panic/recover 경계
      │
      ├── defer recover() → parsePanic → error 변환
      │                   → 그 외 panic → 재던짐 (숨기지 않음)
      ▼
newLexer → newParser → parseExpr → parseTerm → parseFactor
                                                    └── (괄호 만나면) parseExpr 재귀
```

`parsePanic`이 아닌 panic(런타임 에러 등)은 recover하지 않고 그대로 전파된다.

### Lexer

`string` 대신 `[]rune`을 사용하여 멀티바이트 유니코드 문자를 안전하게 처리한다. 공백(`' '`, `'\t'`)은 자동으로 건너뛴다.

---

## 테스트

```bash
go test ./...
go test -race ./...
go vet ./...
```

| 패키지 | 테스트 수 | 내용 |
|---|---|---|
| `internal/calc` | 17개 | Lexer 토큰화, Eval 정상/에러 케이스 |

테스트는 테이블 드리븐(table-driven) 방식으로 작성되었으며, panic 발생 여부는 `recover`를 사용하는 별도 케이스로 검증한다.
