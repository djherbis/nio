package nio

import "io"

type WriterFunc func([]byte) (int, error)

func (w WriterFunc) Write(p []byte) (int, error) {
	return w(p)
}

func (w WriterFunc) Close() error {
	return nil
}

func WriteInChunks(w io.Writer, p []byte, chunk int) (n int, err error) {
	var m int
	for len(p[n:]) > 0 {

		if chunk <= len(p[n:]) {
			m, err = w.Write(p[n : n+chunk])
		} else {
			m, err = w.Write(p[n:])
		}

		n += m

		if err != nil {
			return n, err
		}

	}
	return n, err
}
