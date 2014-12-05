package nio

import "io"

type WriterFunc func([]byte) (int, error)

func (w WriterFunc) Write(p []byte) (int, error) {
	return w(p)
}

func (w WriterFunc) Close() error {
	return nil
}

type chunkWriter struct {
	io.Writer
	chunk int
}

func ChunkWriter(w io.Writer, chunk int) io.Writer {
	return &chunkWriter{
		Writer: w,
		chunk:  chunk,
	}
}

func (w *chunkWriter) Write(p []byte) (n int, err error) {
	var m int
	for len(p[n:]) > 0 {

		if n+w.chunk <= len(p) {
			m, err = w.Writer.Write(p[n : n+w.chunk])
		} else {
			m, err = w.Writer.Write(p[n:])
		}

		n += m

		if err != nil {
			return n, err
		}

	}
	return n, err
}
