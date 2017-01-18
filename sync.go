package nio

import (
	"io"
	"sync"
)

// PipeReader is the read half of the pipe.
type PipeReader struct {
	*bufpipe
}

// CloseWithError closes the reader; subsequent writes to the write half of the pipe will return the error err.
func (r *PipeReader) CloseWithError(err error) error {
	if err == nil {
		err = io.ErrClosedPipe
	}
	r.bufpipe.l.Lock()
	defer r.bufpipe.l.Unlock()
	if r.bufpipe.err == nil {
		r.bufpipe.err = err
		r.bufpipe.c.Signal()
	}
	return nil
}

// Close closes the reader; subsequent writes to the write half of the pipe will return the error io.ErrClosedPipe.
func (r *PipeReader) Close() error {
	return r.CloseWithError(nil)
}

// A PipeWriter is the write half of a pipe.
type PipeWriter struct {
	*bufpipe
}

// CloseWithError closes the writer; once the buffer is empty subsequent reads from the read half of the pipe will return
// no bytes and the error err, or io.EOF if err is nil. CloseWithError always returns nil.
func (w *PipeWriter) CloseWithError(err error) error {
	if err == nil {
		err = io.EOF
	}
	w.bufpipe.l.Lock()
	defer w.bufpipe.l.Unlock()
	if w.bufpipe.err == nil {
		w.bufpipe.err = err
		w.bufpipe.c.Signal()
	}
	return nil
}

// Close closes the writer; once the buffer is empty subsequent reads from the read half of the pipe will return
// no bytes and io.EOF after all the buffer has been read.
func (w *PipeWriter) Close() error {
	return w.CloseWithError(nil)
}

type bufpipe struct {
	rl  sync.Mutex
	wl  sync.Mutex
	l   sync.Mutex
	c   *sync.Cond
	b   Buffer
	err error
}

func newBufferedPipe(buf Buffer) *bufpipe {
	s := &bufpipe{
		b: buf,
	}
	s.c = sync.NewCond(&s.l)
	return s
}

func empty(buf Buffer) bool {
	return buf.Len() == 0
}

func gap(buf Buffer) int64 {
	return buf.Cap() - buf.Len()
}

func (r *PipeReader) Read(p []byte) (n int, err error) {
	r.rl.Lock()
	defer r.rl.Unlock()

	r.l.Lock()
	defer r.c.Signal()
	defer r.l.Unlock()

	for empty(r.b) {
		if r.err != nil {
			return 0, r.err
		}

		r.c.Signal()
		r.c.Wait()
	}

	n, err = r.b.Read(p)
	if err == io.EOF {
		err = nil
	}

	return n, err
}

func (w *PipeWriter) Write(p []byte) (int, error) {
	var m int
	var n, space int64
	var err error
	sliceLen := int64(len(p))

	w.wl.Lock()
	defer w.wl.Unlock()

	w.l.Lock()
	defer w.c.Signal()
	defer w.l.Unlock()

	if w.err != nil {
		return 0, io.ErrClosedPipe
	}

	// while there is data to write
	for writeLen := sliceLen; writeLen > 0; writeLen = sliceLen - n {

		// wait for some buffer space to become available
		for space = gap(w.b); space == 0; space = gap(w.b) {
			w.c.Signal()
			w.c.Wait()
		}
		// space > 0, and locked

		var nn int64
		if space < writeLen {
			// => writeLen - space > 0
			// => (sliceLen - n) - space > 0
			// => sliceLen > n + space
			// nn is safe to use for p[:nn]
			nn = n + space
		} else {
			nn = sliceLen
		}

		m, err = w.b.Write(p[n:nn])
		n += int64(m)
		if err != nil {
			return int(n), err
		}

		// one of two cases has occurred:
		// 1. done writing -> writeLen == 0
		// 2. ran out of buffer space -> gap(w.b) == 0
		// both of these cases are handled at the top of this loop
	}

	return int(n), err
}
