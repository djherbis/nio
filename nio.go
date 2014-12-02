// Package nio provides a few buffered io primitives.
package nio

import "io"

type sdReader struct {
	io.ReadCloser
	in io.ReadCloser
}

func (r *sdReader) Close() error {
	r.in.Close()
	r.ReadCloser.Close()
	return nil
}

type Buffer interface {
	Len() int64
	Cap() int64
	io.ReadWriter
}

func Pipe(buf Buffer) (io.ReadCloser, io.WriteCloser) {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	go func() {
		Copy(outWriter, inReader, buf)
		outWriter.Close()
	}()

	return &sdReader{
		ReadCloser: outReader,
		in:         inReader,
	}, inWriter
}

func Copy(dst io.Writer, src io.Reader, buf Buffer) (n int64, err error) {

	pending := newSync(buf)

	go func() {
		_, err := io.Copy(pending, src)
		pending.CloseWithErr(err)
	}()

	return io.Copy(dst, pending)
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
