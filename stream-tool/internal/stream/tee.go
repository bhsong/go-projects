package stream

import "io"

// Tee returns a Reader that reads from r while copying each read to w.
// Stub: returns r as-is without writing to w.
func Tee(r io.Reader, w io.Writer) io.Reader {
	return io.TeeReader(r, w)
}
