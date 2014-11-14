// Package nio provides a few buffered io primitives.
package nio

import (
	"io"
	"sync"

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

// Pipe works similarly to io.Pipe except that it buffers.
// You may specify the buffers to use for Pipe, or pass none to use the default (see Copy)
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

var pool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return buffer.NewUnboundedBuffer(32*1024, 100*1024*1024)
	},
}

// Copy copies from src to dst until either EOF is reached on src or an error occurs.
// It returns the number of bytes copied, and any any errors while writing to dst.
// It uses a buffer.Buffer which reads from the passed src even while dst.Write is blocking.
// If you pass any buffer.Buffers to Copy, they will be used instead of the default
// scheme which buffers 32KB to memory after which it buffers to <=100MB chunked files.
// Your buffer should have a capacity of at least 32KB (to match io.Copy)
func Copy(dst io.Writer, src io.Reader, buf ...buffer.Buffer) (n int64, err error) {

	if len(buf) == 0 {
		buffer := pool.Get().(buffer.Buffer)
		defer pool.Put(buffer)
		defer buffer.Reset()
		buf = append(buf, buffer)
	}

	pending := buffer.NewSync(buffer.NewMulti(buf...))

	go func() {
		io.Copy(pending, src)
		pending.Done()
	}()

	return io.Copy(dst, pending)
}

// NewReader reads from reader until io.EOF or another error, buffering to buf.
// Reads from the returned ReadCloser will read from the buffer.
// If buf is nil it uses the default (see Copy)
func NewReader(reader io.Reader, buf ...buffer.Buffer) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader, buf...)
		w.Close()
	}()
	return r
}

// NewReadCloser reads from reader until io.EOF or another error, buffering to buf.
// reader is closed automatically when an error is encountered.
// Reads from the returned ReadCloser will read from the buffer.
// If buf is nil it uses the default (see Copy)
func NewReadCloser(reader io.ReadCloser, buf ...buffer.Buffer) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		Copy(w, reader, buf...)
		w.Close()
		reader.Close()
	}()
	return r
}

// NewWriter writes to writer from the buffer. Writes to the returned io.WriteCloser
// will be buffered.
// If buf is nil it uses the default (see Copy)
func NewWriter(writer io.Writer, buf ...buffer.Buffer) io.WriteCloser {
	r, w := io.Pipe()
	go Copy(writer, r, buf...)
	return w
}

// NewWriteCloser writes to writer from the buffer. Writes to the returned io.WriteCloser
// will be buffered. The writer is automatically closed when the returned WriteCloser is closed.
// If buf is nil it uses the default (see Copy)
func NewWriteCloser(writer io.WriteCloser, buf ...buffer.Buffer) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		Copy(writer, r, buf...)
		writer.Close()
	}()
	return w
}
