package nio

import (
	"io"
	"sync"
)

type syncBuf struct {
	done chan struct{}
	l    sync.Mutex
	c    *sync.Cond
	b    Buffer
	err  error
}

func newSync(buf Buffer) *syncBuf {
	s := &syncBuf{
		b:    buf,
		done: make(chan struct{}),
	}
	s.c = sync.NewCond(&s.l)
	return s
}

func (r *syncBuf) CloseWithErr(err error) {
	r.err = err
	r.Close()
	return
}

func (r *syncBuf) Close() error {
	close(r.done)
	r.c.Signal()
	return nil
}

func Empty(buf Buffer) bool {
	return buf.Len() == 0
}

func Gap(buf Buffer) int64 {
	return buf.Cap() - buf.Len()
}

func (r *syncBuf) Read(p []byte) (n int, err error) {
	r.l.Lock()
	defer r.c.Signal()
	defer r.l.Unlock()

	for Empty(r.b) {
		select {
		case <-r.done:
			if r.err != nil {
				return 0, r.err
			}
			return 0, io.EOF
		default:
		}

		r.c.Signal()
		r.c.Wait()
	}

	n, _ = r.b.Read(p)

	return n, nil
}

func (w *syncBuf) Write(p []byte) (n int, err error) {
	w.l.Lock()
	defer w.c.Signal()
	defer w.l.Unlock()

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
