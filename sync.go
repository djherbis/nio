package nio

import (
	"io"
	"sync"
)

type bufPipeReader struct {
	*bufpipe
}

func (r *bufPipeReader) CloseWithErr(err error) error {
	if err == nil {
		err = io.ErrClosedPipe
	}
	r.bufpipe.l.Lock()
	defer r.bufpipe.l.Unlock()
	if r.bufpipe.err == nil {
		r.bufpipe.err = err
		close(r.bufpipe.done)
		r.bufpipe.c.Signal()
	}
	return nil
}

func (r *bufPipeReader) Close() error {
	return r.CloseWithErr(nil)
}

type bufPipeWriter struct {
	*bufpipe
}

func (w *bufPipeWriter) CloseWithErr(err error) error {
	if err == nil {
		err = io.EOF
	}
	w.bufpipe.l.Lock()
	defer w.bufpipe.l.Unlock()
	if w.bufpipe.err == nil {
		w.bufpipe.err = err
		close(w.bufpipe.done)
		w.bufpipe.c.Signal()
	}
	return nil
}

func (w *bufPipeWriter) Close() error {
	return w.CloseWithErr(nil)
}

type bufpipe struct {
	done chan struct{}
	l    sync.Mutex
	c    *sync.Cond
	b    Buffer
	err  error
}

func newBufferedPipe(buf Buffer) *bufpipe {
	s := &bufpipe{
		b:    buf,
		done: make(chan struct{}),
	}
	s.c = sync.NewCond(&s.l)
	return s
}

func Empty(buf Buffer) bool {
	return buf.Len() == 0
}

func Gap(buf Buffer) int64 {
	return buf.Cap() - buf.Len()
}

func (r *bufpipe) Read(p []byte) (n int, err error) {
	r.l.Lock()
	defer r.c.Signal()
	defer r.l.Unlock()

	for Empty(r.b) {
		select {
		case <-r.done:
			return 0, r.err
		default:
		}

		r.c.Signal()
		r.c.Wait()
	}

	n, _ = r.b.Read(p)

	return n, nil
}

func (w *bufpipe) Write(p []byte) (n int, err error) {
	w.l.Lock()
	defer w.c.Signal()
	defer w.l.Unlock()

	select {
	case <-w.done:
		return 0, w.err
	default:
	}

	var m int

	// more data to write
	for len(p[n:]) > 0 {

		// writes too big
		for Gap(w.b) < int64(len(p[n:])) {

			// wait for space
			for Gap(w.b) == 0 {
				w.c.Signal()
				w.c.Wait()
			}

			// chunk write to fill space
			m, err = w.b.Write(p[n : int64(n)+Gap(w.b)])
			n += m
			if err != nil {
				return n, err
			}

			// wait for more space
			w.c.Signal()
			w.c.Wait()

		}

		// check if done
		if len(p[n:]) == 0 {
			return n, nil
		}

		// write
		m, err = w.b.Write(p[n:])
		n += m
		if err != nil {
			return n, err
		}
	}

	return n, err
}
