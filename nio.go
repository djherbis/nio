// Package nio provides a few buffered io primitives.
package nio

import "io"

type Buffer interface {
	Len() int64
	Cap() int64
	io.ReadWriter
}

func Pipe(buf Buffer) (r io.ReadCloser, w io.WriteCloser) {
	p := newBufferedPipe(buf)
	r = &bufPipeReader{bufpipe: p}
	w = &bufPipeWriter{bufpipe: p}
	return r, w
}

func Copy(dst io.Writer, src io.Reader, buf Buffer) (n int64, err error) {
	r, w := Pipe(buf)

	go func() {
		_, err := io.Copy(w, src)
		w.(*bufPipeWriter).CloseWithErr(err)
	}()

	return io.Copy(dst, r)
}

func NewReader(reader io.Reader, buf Buffer) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader, buf)
		w.Close()
	}()
	return r
}

func NewReadCloser(reader io.ReadCloser, buf Buffer) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader, buf)
		w.Close()
		reader.Close()
	}()
	return r
}

func NewWriter(writer io.Writer, buf Buffer) io.WriteCloser {
	r, w := io.Pipe()
	go Copy(writer, r, buf)
	return w
}

func NewWriteCloser(writer io.WriteCloser, buf Buffer) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		Copy(writer, r, buf)
		writer.Close()
	}()
	return w
}
