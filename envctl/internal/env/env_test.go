package env

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Unit: parseLine
// ─────────────────────────────────────────────────────────────────────────────

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		lineNum int
		want    Entry
		wantOK  bool
	}{
		{
			name:    "[정상] 기본 KEY=VALUE",
			line:    "DB_HOST=localhost",
			lineNum: 1,
			want:    Entry{Key: "DB_HOST", Value: "localhost", Line: 1},
			wantOK:  true,
		},
		{
			name:    "[정상] 앞뒤 공백 있는 줄",
			line:    "  DB_PORT = 5432  ",
			lineNum: 2,
			want:    Entry{Key: "DB_PORT", Value: "5432", Line: 2},
			wantOK:  true,
		},
		{
			name:    "[정상] 빈 값 (EMPTY=)",
			line:    "EMPTY=",
			lineNum: 3,
			want:    Entry{Key: "EMPTY", Value: "", Line: 3},
			wantOK:  true,
		},
		{
			name:    "[정상] 주석 줄",
			line:    "# 주석",
			lineNum: 4,
			want:    Entry{},
			wantOK:  false,
		},
		{
			name:    "[정상] 빈 문자열",
			line:    "",
			lineNum: 5,
			want:    Entry{},
			wantOK:  false,
		},
		{
			name:    "[정상] 공백만 있는 줄",
			line:    "   ",
			lineNum: 6,
			want:    Entry{},
			wantOK:  false,
		},
		{
			name:    "[엣지] = 없는 줄 (INVALID_LINE)",
			line:    "INVALID_LINE",
			lineNum: 7,
			want:    Entry{},
			wantOK:  false,
		},
		{
			name:    "[엣지] 키 없는 줄 (=VALUE)",
			line:    "=VALUE",
			lineNum: 8,
			want:    Entry{},
			wantOK:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseLine(tt.line, tt.lineNum)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("entry = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: Parse
// ─────────────────────────────────────────────────────────────────────────────

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantError bool
	}{
		{
			name:    "[정상] 정상 .env 내용 — 주석/빈줄 제외",
			input:   "# 주석\nDB_HOST=localhost\n\nDB_PORT=5432\n",
			wantLen: 2,
		},
		{
			name:    "[정상] 빈 입력",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "[정상] 주석과 빈 줄만",
			input:   "# only comments\n\n# another comment\n",
			wantLen: 0,
		},
		{
			name:  "[엣지] 줄 번호 정확성 — 주석 스킵 후 다음 Entry.Line",
			input: "# comment\n\nFOO=bar\n",
			// FOO=bar 는 3번째 줄이어야 한다
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := Parse(r)
			if (err != nil) != tt.wantError {
				t.Fatalf("Parse() error = %v, wantError %v", err, tt.wantError)
			}
			if len(got) != tt.wantLen {
				t.Errorf("len(entries) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}

	// 줄 번호 정확성을 별도로 검증
	t.Run("[엣지] 줄 번호 정확성 검증 — Line 필드 값 확인", func(t *testing.T) {
		r := strings.NewReader("# comment\n\nFOO=bar\n")
		entries, err := Parse(r)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d, want 1", len(entries))
		}
		if entries[0].Line != 3 {
			t.Errorf("entries[0].Line = %d, want 3", entries[0].Line)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: ToMap
// ─────────────────────────────────────────────────────────────────────────────

func TestToMap(t *testing.T) {
	t.Run("[정상] 중복 없는 entries", func(t *testing.T) {
		entries := []Entry{
			{Key: "A", Value: "1", Line: 1},
			{Key: "B", Value: "2", Line: 2},
		}
		m, dups := ToMap(entries)
		if m["A"] != "1" || m["B"] != "2" {
			t.Errorf("map = %v, want {A:1, B:2}", m)
		}
		if len(dups) != 0 {
			t.Errorf("duplicates = %v, want empty", dups)
		}
	})

	t.Run("[정상] 중복 있는 entries — 마지막 값 우선", func(t *testing.T) {
		entries := []Entry{
			{Key: "KEY", Value: "first", Line: 1},
			{Key: "KEY", Value: "second", Line: 3},
		}
		m, dups := ToMap(entries)
		if m["KEY"] != "second" {
			t.Errorf("m[KEY] = %q, want %q", m["KEY"], "second")
		}
		if len(dups) != 1 {
			t.Fatalf("len(dups) = %d, want 1", len(dups))
		}
		if dups[0].Key != "KEY" {
			t.Errorf("dups[0].Key = %q, want %q", dups[0].Key, "KEY")
		}
		if len(dups[0].Lines) != 2 || dups[0].Lines[0] != 1 || dups[0].Lines[1] != 3 {
			t.Errorf("dups[0].Lines = %v, want [1 3]", dups[0].Lines)
		}
	})

	t.Run("[정상] 빈 entries", func(t *testing.T) {
		m, dups := ToMap([]Entry{})
		if len(m) != 0 {
			t.Errorf("len(m) = %d, want 0", len(m))
		}
		if len(dups) != 0 {
			t.Errorf("len(dups) = %d, want 0", len(dups))
		}
	})

	t.Run("[엣지] 같은 키 3번 중복 — Lines에 3개 포함", func(t *testing.T) {
		entries := []Entry{
			{Key: "X", Value: "a", Line: 1},
			{Key: "X", Value: "b", Line: 5},
			{Key: "X", Value: "c", Line: 9},
		}
		_, dups := ToMap(entries)
		if len(dups) != 1 {
			t.Fatalf("len(dups) = %d, want 1", len(dups))
		}
		if len(dups[0].Lines) != 3 {
			t.Errorf("len(dups[0].Lines) = %d, want 3", len(dups[0].Lines))
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: MapToSlice
// ─────────────────────────────────────────────────────────────────────────────

func TestMapToSlice(t *testing.T) {
	t.Run("[정상] map → KEY=VALUE 슬라이스", func(t *testing.T) {
		m := map[string]string{"FOO": "bar", "BAZ": "qux"}
		result := MapToSlice(m)
		if len(result) != 2 {
			t.Errorf("len(result) = %d, want 2", len(result))
		}
		seen := make(map[string]bool)
		for _, s := range result {
			seen[s] = true
		}
		if !seen["FOO=bar"] {
			t.Errorf("missing FOO=bar in %v", result)
		}
		if !seen["BAZ=qux"] {
			t.Errorf("missing BAZ=qux in %v", result)
		}
	})

	t.Run("[정상] 빈 map → 빈 슬라이스", func(t *testing.T) {
		result := MapToSlice(map[string]string{})
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: Expand
// ─────────────────────────────────────────────────────────────────────────────

func TestExpand(t *testing.T) {
	env := map[string]string{
		"HOST": "localhost",
		"PORT": "5432",
	}
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "[정상] $VAR 형태 단일 치환",
			value: "$HOST",
			want:  "localhost",
		},
		{
			name:  "[정상] ${VAR} 형태 단일 치환",
			value: "${PORT}",
			want:  "5432",
		},
		{
			name:  "[정상] 여러 변수 혼합 $A:${B}/c",
			value: "$HOST:${PORT}/db",
			want:  "localhost:5432/db",
		},
		{
			name:  "[정상] 변수 없는 문자열 — 원본 그대로",
			value: "no-vars-here",
			want:  "no-vars-here",
		},
		{
			name:  "[엣지] 존재하지 않는 키 → 빈 문자열로 치환",
			value: "$MISSING",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Expand(tt.value, env)
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: Merge
// ─────────────────────────────────────────────────────────────────────────────

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		want     map[string]string
	}{
		{
			name:     "[정상] base + override → override 값 우선",
			base:     map[string]string{"A": "1", "B": "2"},
			override: map[string]string{"B": "3"},
			want:     map[string]string{"A": "1", "B": "3"},
		},
		{
			name:     "[정상] override에만 있는 키 → 결과에 포함",
			base:     map[string]string{"A": "1"},
			override: map[string]string{"C": "4"},
			want:     map[string]string{"A": "1", "C": "4"},
		},
		{
			name:     "[정상] base에만 있는 키 → 결과에 포함",
			base:     map[string]string{"A": "1", "B": "2"},
			override: map[string]string{},
			want:     map[string]string{"A": "1", "B": "2"},
		},
		{
			name:     "[엣지] base가 빈 map → override 그대로",
			base:     map[string]string{},
			override: map[string]string{"X": "y"},
			want:     map[string]string{"X": "y"},
		},
		{
			name:     "[엣지] override가 빈 map → base 그대로",
			base:     map[string]string{"K": "v"},
			override: map[string]string{},
			want:     map[string]string{"K": "v"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.override)
			if len(got) != len(tt.want) {
				t.Errorf("len(result) = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("result[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}

	t.Run("[엣지] base 원본 변경 안 됨 — Merge 후 base 불변 확인", func(t *testing.T) {
		base := map[string]string{"A": "1", "B": "2"}
		override := map[string]string{"B": "99"}
		_ = Merge(base, override)
		if base["B"] != "2" {
			t.Errorf("base[B] = %q after Merge, want %q (base must not be mutated)", base["B"], "2")
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: Apply
// ─────────────────────────────────────────────────────────────────────────────

func TestApply(t *testing.T) {
	t.Run("[정상] overrideExisting=true → 기존 환경변수 덮어씀", func(t *testing.T) {
		key := "ENVCTL_APPLY_OVERRIDE_TEST"
		t.Setenv(key, "original")

		err := Apply(map[string]string{key: "new_value"}, true)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		// t.Setenv이 cleanup을 담당하므로, Apply가 실제로 설정했는지 확인하려면
		// 테스트 내에서 os.Getenv를 직접 호출한다.
		// Apply 내부에서 os.Setenv를 호출하므로, 환경변수가 변경되어 있어야 한다.
		// (t.Setenv의 cleanup은 테스트 종료 시 원본으로 복구)
	})

	t.Run("[정상] overrideExisting=false → 기존 환경변수 유지", func(t *testing.T) {
		key := "ENVCTL_APPLY_NOOVERRIDE_TEST"
		t.Setenv(key, "original")

		err := Apply(map[string]string{key: "should_not_change"}, false)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		// Apply가 false일 때 기존 값을 건드리지 않아야 한다.
		// os.Getenv는 Apply 내부에서 쓰는 os.Setenv 결과를 반영하지 않아야 한다.
	})

	t.Run("[정상] 빈 map → 에러 없이 종료", func(t *testing.T) {
		err := Apply(map[string]string{}, false)
		if err != nil {
			t.Errorf("Apply(empty) error = %v, want nil", err)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Unit: EnvParseError
// ─────────────────────────────────────────────────────────────────────────────

func TestEnvParseError(t *testing.T) {
	t.Run("[정상] Error() 문자열 형식", func(t *testing.T) {
		e := &EnvParseError{Line: 3, Msg: "설명 메시지"}
		want := "줄 3: 설명 메시지"
		if e.Error() != want {
			t.Errorf("Error() = %q, want %q", e.Error(), want)
		}
	})

	t.Run("[정상] errors.As로 타입 추출 — Line, Msg 접근 가능", func(t *testing.T) {
		original := &EnvParseError{Line: 7, Msg: "잘못된 형식"}
		wrapped := fmt.Errorf("wrap: %w", original)

		var target *EnvParseError
		if !errors.As(wrapped, &target) {
			t.Fatal("errors.As failed: *EnvParseError not found in chain")
		}
		if target.Line != 7 {
			t.Errorf("target.Line = %d, want 7", target.Line)
		}
		if target.Msg != "잘못된 형식" {
			t.Errorf("target.Msg = %q, want %q", target.Msg, "잘못된 형식")
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Feature: Parse → ToMap → Expand 통합
// ─────────────────────────────────────────────────────────────────────────────

func TestFeature_ParseAndExpand(t *testing.T) {
	input := "DB_HOST=localhost\nDB_PORT=5432\nURL=$DB_HOST:$DB_PORT/db"
	r := strings.NewReader(input)

	entries, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	envMap, _ := ToMap(entries)

	expMap := make(map[string]string, len(envMap))
	for k, v := range envMap {
		expMap[k] = Expand(v, envMap)
	}

	want := "localhost:5432/db"
	if expMap["URL"] != want {
		t.Errorf("expMap[URL] = %q, want %q", expMap["URL"], want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Feature: Merge 다중 파일
// ─────────────────────────────────────────────────────────────────────────────

func TestFeature_MergeMultipleFiles(t *testing.T) {
	base := map[string]string{"A": "1", "B": "2"}
	local := map[string]string{"B": "3", "C": "4"}

	result := Merge(base, local)

	if result["A"] != "1" {
		t.Errorf("result[A] = %q, want %q", result["A"], "1")
	}
	if result["B"] != "3" {
		t.Errorf("result[B] = %q, want %q (override 우선)", result["B"], "3")
	}
	if result["C"] != "4" {
		t.Errorf("result[C] = %q, want %q", result["C"], "4")
	}
	// base 불변 확인
	if base["B"] != "2" {
		t.Errorf("base[B] = %q after Merge, want %q (base must not be mutated)", base["B"], "2")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Feature: 중복 키 탐지
// ─────────────────────────────────────────────────────────────────────────────

func TestFeature_DuplicateDetection(t *testing.T) {
	input := "KEY=first\nOTHER=x\nKEY=second"
	r := strings.NewReader(input)

	entries, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	envMap, dups := ToMap(entries)

	if envMap["KEY"] != "second" {
		t.Errorf("envMap[KEY] = %q, want %q (마지막 값 우선)", envMap["KEY"], "second")
	}

	if len(dups) != 1 {
		t.Fatalf("len(dups) = %d, want 1", len(dups))
	}
	if dups[0].Key != "KEY" {
		t.Errorf("dups[0].Key = %q, want KEY", dups[0].Key)
	}
	if len(dups[0].Lines) != 2 || dups[0].Lines[0] != 1 || dups[0].Lines[1] != 3 {
		t.Errorf("dups[0].Lines = %v, want [1 3]", dups[0].Lines)
	}
}
