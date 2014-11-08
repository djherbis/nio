package nio

import (
	"io"

	"github.com/djherbis/buffer"
)

type sdReader struct {
	io.ReadCloser
	in    io.ReadCloser
	chout chan interface{}
}

func (r *sdReader) Close() error {
	r.in.Close()
	r.ReadCloser.Close()
	for _ = range r.chout {
	}
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
