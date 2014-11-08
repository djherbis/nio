package nio

import (
	"io"

	"github.com/djherbis/buffer"
)

func NewReader(reader io.Reader, buf ...buffer.Buffer) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader, buf...)
		w.Close()
	}()
	return r
}

func NewReadCloser(reader io.ReadCloser, buf ...buffer.Buffer) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader)
		w.Close()
		reader.Close()
	}()
	return r
}

func NewWriter(writer io.Writer, buf ...buffer.Buffer) io.WriteCloser {
	r, w := io.Pipe()
	go Copy(writer, r, buf...)
	return w
}

func NewWriteCloser(writer io.WriteCloser, buf ...buffer.Buffer) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		Copy(writer, r, buf...)
		writer.Close()
	}()
	return w
}
