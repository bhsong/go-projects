package stream

import "io"

func MultiWrite(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}
