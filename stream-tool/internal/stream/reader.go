package stream

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Stats struct {
	Lines int64
	Words int64
	Bytes int64
}

type CountingReader struct {
	r     io.Reader
	Bytes int64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{
		r: r,
	}
}

func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	if n > 0 {
		cr.Bytes += int64(n)
	}
	return n, err
}

func Count(r io.Reader) (Stats, error) {

	scanner := bufio.NewScanner(r)

	var lines, words, bytes int64

	for scanner.Scan() {
		line := scanner.Text()
		lines += 1

		fields := strings.Fields(line)
		words += int64(len(fields))
		bytes += int64(len(line)) + 1
	}

	if err := scanner.Err(); err != nil {
		return Stats{}, fmt.Errorf("stream.Count: read error: %w", err)
	}

	return Stats{lines, words, bytes}, nil
}

var _ io.Reader = (*CountingReader)(nil)
