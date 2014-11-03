package nio

import "io"

func NewReader(reader io.Reader) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader)
		w.Close()
	}()
	return r
}

func NewReadCloser(reader io.ReadCloser) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader)
		w.Close()
		reader.Close()
	}()
	return r
}

func NewWriter(writer io.Writer) io.WriteCloser {
	r, w := io.Pipe()
	go Copy(writer, r)
	return w
}

func NewWriteCloser(writer io.WriteCloser) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		Copy(writer, r)
		writer.Close()
	}()
	return w
}
