// Package nio provides a few buffered io primitives.
package nio

import "io"

type Buffer interface {
	Len() int64
	Cap() int64
	io.ReadWriter
}

// Pipe creates a buffered pipe.
// It can be used to connect code expecting an io.Reader with code expecting an io.Writer.
// Reads on one end read from the supplied Buffer. Writes write to the supplied Buffer.
// It is safe to call Read and Write in parallel with each other or with Close.
// Close will complete once pending I/O is done. Buffered data will still be available to Read
// after the Writer has been closed. Parallel calls to Read, and parallel calls to Write are also safe :
// the individual calls will be gated sequentially.
func Pipe(buf Buffer) (r io.ReadCloser, w io.WriteCloser) {
	p := newBufferedPipe(buf)
	r = &bufPipeReader{bufpipe: p}
	w = &bufPipeWriter{bufpipe: p}
	return r, w
}

// Copy copies from src to buf, and from buf to dst in parallel until
// either EOF is reached on src or an error occurs. It returns the number of bytes
// copied to dst and the first error encountered while copying, if any.
// EOF is not considered to be an error. If src implements WriterTo, it is used to
// write to the supplied Buffer. If dst implements ReaderFrom, it is used to read from
// the supplied Buffer.
func Copy(dst io.Writer, src io.Reader, buf Buffer) (n int64, err error) {
	return io.Copy(dst, NewReader(src, buf))
}

func NewReader(src io.Reader, buf Buffer) io.ReadCloser {
	r, w := Pipe(buf)

	go func() {
		_, err := io.Copy(w, src)
		w.(*bufPipeWriter).CloseWithErr(err)
	}()

	return r
}
