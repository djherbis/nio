// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Pipe adapter to connect code expecting an io.Reader
// with code expecting an io.Writer.

package nio

import (
	"errors"
	"io"
	"sync"
	"time"
)

// ErrClosedPipe is the error used for read or write operations on a closed pipe.
var ErrClosedPipe = errors.New("io: read/write on closed pipe")
var ErrReadTimeout = errors.New("read timeout")
var ErrWriteTimeout = errors.New("write timeout")

type pipeResult struct {
	n   int
	err error
}

// A pipe is the shared pipe structure underlying PipeReader and PipeWriter.
type pipe struct {
	rl    sync.Mutex    // gates readers one at a time
	wl    sync.Mutex    // gates writers one at a time
	l     sync.Mutex    // protects remaining fields
	rt    *time.Timer   // read timeout
	wt    *time.Timer   // write timeout
	rdt   time.Time     // read deadline
	wdt   time.Time     // write deadline
	rtc   chan struct{} // read timeout channel
	wtc   chan struct{} // write timeout channel
	data  []byte        // data remaining in pending write
	rwait sync.Cond     // waiting reader
	wwait sync.Cond     // waiting writer
	rerr  error         // if reader closed, error to give writes
	werr  error         // if writer closed, error to give reads
}

func (p *pipe) readTimeout() {
	if !p.rdt.IsZero() {
		if p.rt == nil {
			p.rt = time.AfterFunc(p.rdt.Sub(time.Now()), func() {
				p.rtc <- struct{}{}
				p.rwait.Signal()
			})
		} else {
			p.rt.Reset(p.rdt.Sub(time.Now()))
		}
	}
}

func (p *pipe) read(b []byte) (n int, err error) {
	// One reader at a time.
	p.rl.Lock()
	defer p.rl.Unlock()

	p.l.Lock()
	defer p.l.Unlock()

	p.readTimeout()

	for {
		select {
		case <-p.rtc:
			// BUG(Dustin): Clear state.
			return 0, ErrReadTimeout
		default:
		}

		if p.rerr != nil {
			return 0, ErrClosedPipe
		}
		if p.data != nil {
			break
		}
		if p.werr != nil {
			return 0, p.werr
		}
		p.rwait.Wait()
	}
	n = copy(b, p.data)
	p.data = p.data[n:]
	if len(p.data) == 0 {
		p.data = nil
		if p.wt != nil {
			p.wt.Stop()
		}
		p.wwait.Signal()
	}
	return
}

var zero [0]byte

func (p *pipe) writeTimeout() {
	if !p.wdt.IsZero() {
		if p.wt == nil {
			p.wt = time.AfterFunc(p.wdt.Sub(time.Now()), func() {
				p.wtc <- struct{}{}
				p.wwait.Signal()
			})
		} else {
			p.wt.Reset(p.wdt.Sub(time.Now()))
		}
	}
}

func (p *pipe) write(b []byte) (n int, err error) {
	// pipe uses nil to mean not available
	if b == nil {
		b = zero[:]
	}

	// One writer at a time.
	p.wl.Lock()
	defer p.wl.Unlock()

	p.l.Lock()
	defer p.l.Unlock()

	p.writeTimeout()

	if p.werr != nil {
		err = ErrClosedPipe
		return
	}
	p.data = b
	if p.rt != nil {
		p.rt.Stop()
	}
	p.rwait.Signal()
	for {
		select {
		case <-p.wtc:
			// BUG(Dustin): Clear state.
			n = len(b) - len(p.data)
			p.data = nil
			return n, ErrWriteTimeout
		default:
		}

		if p.data == nil {
			break
		}
		if p.rerr != nil {
			err = p.rerr
			break
		}
		if p.werr != nil {
			err = ErrClosedPipe
		}
		p.wwait.Wait()
	}
	n = len(b) - len(p.data)
	p.data = nil // in case of rerr or werr
	return
}

func (p *pipe) rclose(err error) {
	if err == nil {
		err = ErrClosedPipe
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.rerr = err
	p.rwait.Signal()
	p.wwait.Signal()
}

func (p *pipe) wclose(err error) {
	if err == nil {
		err = io.EOF
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.werr = err
	p.rwait.Signal()
	p.wwait.Signal()
}

// A PipeReader is the read half of a pipe.
type PipeReader struct {
	p *pipe
}

func (w *PipeReader) SetReadDeadline(t time.Time) error {
	w.p.l.Lock()
	defer w.p.l.Unlock()
	w.p.rdt = t
	if w.p.rt != nil {
		w.p.rt.Reset(t.Sub(time.Now()))
	}
	return nil
}

// Read implements the standard Read interface:
// it reads data from the pipe, blocking until a writer
// arrives or the write end is closed.
// If the write end is closed with an error, that error is
// returned as err; otherwise err is EOF.
func (r *PipeReader) Read(data []byte) (n int, err error) {
	return r.p.read(data)
}

// Close closes the reader; subsequent writes to the
// write half of the pipe will return the error ErrClosedPipe.
func (r *PipeReader) Close() error {
	return r.CloseWithError(nil)
}

// CloseWithError closes the reader; subsequent writes
// to the write half of the pipe will return the error err.
func (r *PipeReader) CloseWithError(err error) error {
	r.p.rclose(err)
	return nil
}

// A PipeWriter is the write half of a pipe.
type PipeWriter struct {
	p *pipe
}

func (w *PipeWriter) SetWriteDeadline(t time.Time) error {
	w.p.l.Lock()
	defer w.p.l.Unlock()
	w.p.wdt = t
	if w.p.wt != nil {
		w.p.wt.Reset(t.Sub(time.Now()))
	}
	return nil
}

// Write implements the standard Write interface:
// it writes data to the pipe, blocking until readers
// have consumed all the data or the read end is closed.
// If the read end is closed with an error, that err is
// returned as err; otherwise err is ErrClosedPipe.
func (w *PipeWriter) Write(data []byte) (n int, err error) {
	return w.p.write(data)
}

// Close closes the writer; subsequent reads from the
// read half of the pipe will return no bytes and EOF.
func (w *PipeWriter) Close() error {
	return w.CloseWithError(nil)
}

// CloseWithError closes the writer; subsequent reads from the
// read half of the pipe will return no bytes and the error err.
func (w *PipeWriter) CloseWithError(err error) error {
	w.p.wclose(err)
	return nil
}

// TimeoutPipe creates a synchronous in-memory pipe with optional Read/Write Deadlines.
// It can be used to connect code expecting an io.Reader
// with code expecting an io.Writer.
// Reads on one end are matched with writes on the other,
// copying data directly between the two; there is no internal buffering.
// It is safe to call Read and Write in parallel with each other or with
// Close. Close will complete once pending I/O is done. Parallel calls to
// Read, and parallel calls to Write, are also safe:
// the individual calls will be gated sequentially.
func TimeoutPipe() (*PipeReader, *PipeWriter) {
	p := new(pipe)
	p.rwait.L = &p.l
	p.wwait.L = &p.l
	p.rtc = make(chan struct{}, 1)
	p.wtc = make(chan struct{}, 1)
	r := &PipeReader{p}
	w := &PipeWriter{p}
	return r, w
}
