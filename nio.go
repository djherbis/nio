package nio

import (
	"io"

	"github.com/djherbis/buffer"
)

type sdReader struct {
	io.ReadCloser
	in io.ReadCloser
}

func (r *sdReader) Close() error {
	r.in.Close()
	r.ReadCloser.Close()
	return nil
}

func Pipe(buf ...buffer.Buffer) (io.ReadCloser, io.WriteCloser) {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	go func() {
		Copy(outWriter, inReader, buf...)
		outWriter.Close()
	}()

	return &sdReader{
		ReadCloser: outReader,
		in:         inReader,
	}, inWriter
}

func Copy(dst io.Writer, src io.Reader, buf ...buffer.Buffer) (n int64, err error) {

	if len(buf) == 0 {
		buf = append(buf, buffer.NewUnboundedBuffer(32*1024, 100*1024*1024))
	}

	pending := buffer.NewSync(buffer.NewMulti(buf...))

	go func() {
		io.Copy(pending, src)
		pending.Done()
	}()

	return io.Copy(dst, pending)
}

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
		Copy(w, reader, buf...)
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
