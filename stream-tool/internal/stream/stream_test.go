package stream_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/bhsong/go-projects/stream-tool/internal/stream"
)

// ── Unit Tests ────────────────────────────────────────────────────────────────

func TestCountingReader(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantBytes int64
	}{
		{
			name:      "[정상] 짧은 문자열 읽기 — Bytes가 실제 바이트 수와 일치",
			input:     "hello",
			wantBytes: 5,
		},
		{
			name:      "[정상] 여러 번 Read 호출 — Bytes가 누적됨",
			input:     "hello world!",
			wantBytes: 12,
		},
		{
			name:      "[엣지] 빈 Reader — Bytes == 0",
			input:     "",
			wantBytes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := stream.NewCountingReader(strings.NewReader(tt.input))
			// 1-byte 버퍼로 Read를 여러 번 호출해 누적 동작을 검증한다.
			buf := make([]byte, 1)
			for {
				_, err := cr.Read(buf)
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatalf("cr.Read: %v", err)
				}
			}
			if cr.Bytes != tt.wantBytes {
				t.Errorf("Bytes = %d, want %d", cr.Bytes, tt.wantBytes)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLines int64
		wantWords int64
		wantBytes int64
		wantErr   bool
	}{
		{
			// scanner.Text()="hello world" (len=11), Bytes += 11+1 = 12
			name:      "[정상] 단일 줄",
			input:     "hello world\n",
			wantLines: 1,
			wantWords: 2,
			wantBytes: 12,
		},
		{
			// 12 + 12 = 24
			name:      "[정상] 여러 줄 텍스트 — Lines/Words/Bytes 정확히 집계",
			input:     "hello world\nfoo bar baz\n",
			wantLines: 2,
			wantWords: 5,
			wantBytes: 24,
		},
		{
			// "hello"(6) + ""(1) + "world"(6) = 13
			name:      "[정상] 빈 줄 포함 텍스트 — 빈 줄도 Lines에 포함, Words는 0 추가",
			input:     "hello\n\nworld\n",
			wantLines: 3,
			wantWords: 2,
			wantBytes: 13,
		},
		{
			name:      "[엣지] 빈 Reader — Stats{0,0,0} 반환, error nil",
			input:     "",
			wantLines: 0,
			wantWords: 0,
			wantBytes: 0,
		},
		{
			// 개행 없는 마지막 줄은 scanner.Scan()이 반환 → Lines:1. Bytes는 +1 보정으로 6.
			name:      "[엣지] 개행 없는 마지막 줄 — Lines:1 반환",
			input:     "hello",
			wantLines: 1,
			wantWords: 1,
			wantBytes: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := stream.Count(strings.NewReader(tt.input))
			if tt.wantErr && err == nil {
				t.Fatal("want error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stats.Lines != tt.wantLines {
				t.Errorf("Lines = %d, want %d", stats.Lines, tt.wantLines)
			}
			if stats.Words != tt.wantWords {
				t.Errorf("Words = %d, want %d", stats.Words, tt.wantWords)
			}
			if stats.Bytes != tt.wantBytes {
				t.Errorf("Bytes = %d, want %d", stats.Bytes, tt.wantBytes)
			}
		})
	}
}

func TestTee(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		doRead  bool   // TeeReader에서 실제로 읽을지 여부
		wantBuf string // 읽은 후 w에 기록되어야 할 내용
	}{
		{
			name:    "[정상] r에서 읽으면 w에도 동일 내용이 기록됨",
			input:   "hello, tee",
			doRead:  true,
			wantBuf: "hello, tee",
		},
		{
			name:    "[정상] 읽기 전에는 w에 아무것도 쓰이지 않음",
			input:   "hello, tee",
			doRead:  false,
			wantBuf: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			teeReader := stream.Tee(strings.NewReader(tt.input), &buf)
			if tt.doRead {
				if _, err := io.ReadAll(teeReader); err != nil {
					t.Fatalf("ReadAll: %v", err)
				}
			}
			if got := buf.String(); got != tt.wantBuf {
				t.Errorf("buf = %q, want %q", got, tt.wantBuf)
			}
		})
	}
}

// errWriter는 Write 호출 시 항상 에러를 반환한다.
type errWriter struct{}

func (e errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("forced write error")
}

func TestMultiWrite(t *testing.T) {
	tests := []struct {
		name string
		// 각 케이스마다 신선한 Writer를 생성하기 위해 팩토리 함수를 사용한다.
		makeWriters func() []io.Writer
		input       string
		wantErr     bool
	}{
		{
			name: "[정상] 두 Writer에 동일 내용이 기록됨",
			makeWriters: func() []io.Writer {
				return []io.Writer{new(bytes.Buffer), new(bytes.Buffer)}
			},
			input: "hello",
		},
		{
			name: "[정상] 세 Writer 모두에 기록됨",
			makeWriters: func() []io.Writer {
				return []io.Writer{new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)}
			},
			input: "hello world",
		},
		{
			// errWriter를 두 번째에 배치해 스텁(첫 번째만 반환)이 우연히 통과하는 것을 방지한다.
			name: "[엣지] Writer 중 하나가 실패 — 에러 반환",
			makeWriters: func() []io.Writer {
				return []io.Writer{new(bytes.Buffer), errWriter{}}
			},
			input:   "hello",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := tt.makeWriters()
			mw := stream.MultiWrite(ws...)
			_, err := mw.Write([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Write: %v", err)
			}
			// 모든 Buffer Writer에 동일한 내용이 기록되어야 한다.
			for i, w := range ws {
				buf, ok := w.(*bytes.Buffer)
				if !ok {
					continue
				}
				if got := buf.String(); got != tt.input {
					t.Errorf("writers[%d] = %q, want %q", i, got, tt.input)
				}
			}
		})
	}
}

// ── Feature Tests ─────────────────────────────────────────────────────────────

func TestFeature_TeeAndCount(t *testing.T) {
	input := "hello world\nfoo bar baz\n"

	var teeBuf bytes.Buffer
	teeReader := stream.Tee(strings.NewReader(input), &teeBuf)
	stats, err := stream.Count(teeReader)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}

	if stats.Lines != 2 {
		t.Errorf("Lines = %d, want 2", stats.Lines)
	}
	if stats.Words != 5 {
		t.Errorf("Words = %d, want 5", stats.Words)
	}
	if stats.Bytes != 24 {
		t.Errorf("Bytes = %d, want 24", stats.Bytes)
	}
	// tee 버퍼는 원본 텍스트를 그대로 보존해야 한다.
	if got := teeBuf.String(); got != input {
		t.Errorf("teeBuf = %q, want %q", got, input)
	}
}

func TestFeature_MultiWriteAndCount(t *testing.T) {
	input := "hello world\nfoo bar baz\n"

	stats, err := stream.Count(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Count: %v", err)
	}

	var buf1, buf2 bytes.Buffer
	mw := stream.MultiWrite(&buf1, &buf2)
	fmt.Fprintf(mw, "줄: %d\n단어: %d\n바이트: %d\n", stats.Lines, stats.Words, stats.Bytes)

	if buf1.String() != buf2.String() {
		t.Errorf("buf1 = %q, buf2 = %q: want identical", buf1.String(), buf2.String())
	}
	if buf1.String() == "" {
		t.Error("output must not be empty")
	}
}

func TestFeature_CountingReaderWithCount(t *testing.T) {
	input := "hello world\nfoo bar baz\n"

	cr := stream.NewCountingReader(strings.NewReader(input))
	stats, err := stream.Count(cr)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}

	// CountingReader.Bytes는 Count가 집계한 Stats.Bytes와 같아야 한다.
	if cr.Bytes != stats.Bytes {
		t.Errorf("cr.Bytes = %d, stats.Bytes = %d: want equal", cr.Bytes, stats.Bytes)
	}
	// 빈 입력이 아니므로 Bytes > 0 이어야 한다.
	if stats.Bytes == 0 {
		t.Error("stats.Bytes must be > 0 for non-empty input")
	}
}
