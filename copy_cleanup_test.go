package nio

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/djherbis/buffer"
)

// Observe the standard bytes.Reader fast path without replacing its transfer.
type observedCopySource struct {
	*bytes.Reader
	done   chan struct{}
	writer chan io.Writer
	closed bool
}

func (s *observedCopySource) WriteTo(w io.Writer) (int64, error) {
	s.writer <- w
	defer close(s.done)
	return s.Reader.WriteTo(w)
}

func (s *observedCopySource) Close() error { s.closed = true; return nil }

type copyErrorWriter struct {
	n   int
	err error
}

func (w copyErrorWriter) Write(p []byte) (int, error) { return w.n, w.err }

func TestCopyDestinationErrorStopsProducer(t *testing.T) {
	writeErr := errors.New("destination failed")
	closedConn, peer := net.Pipe()
	defer closedConn.Close()
	peer.Close()

	for _, tc := range []struct {
		name    string
		dst     io.Writer
		wantN   int64
		wantErr error
	}{
		{"closed connection", closedConn, 0, io.ErrClosedPipe},
		{"partial write with error", copyErrorWriter{7, writeErr}, 7, writeErr},
		{"short write", copyErrorWriter{7, nil}, 7, io.ErrShortWrite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The source is much larger than the buffer, so an early destination
			// failure leaves its WriterTo blocked in a pipe write unless Copy
			// closes the read end it owns.
			src := &observedCopySource{
				Reader: bytes.NewReader(bytes.Repeat([]byte("x"), 64*1024)),
				done:   make(chan struct{}),
				writer: make(chan io.Writer, 1),
			}
			buf := buffer.New(32)
			n, err := Copy(tc.dst, src, buf)
			if n != tc.wantN || err != tc.wantErr {
				t.Errorf("Copy() = %d, %v; want %d, %v", n, err, tc.wantN, tc.wantErr)
			}
			// Copy has returned, so its caller can reset the supplied buffer.
			buf.Reset()
			select {
			case <-src.done:
			case <-time.After(time.Second):
				// Also clean up the producer when run against the broken base.
				(<-src.writer).(io.Closer).Close()
				<-src.done
				t.Fatal("Copy returned but its producer is still blocked")
			}

			if src.closed {
				t.Error("Copy closed the caller-owned source")
			}

			// The supplied buffer can then serve another complete transfer.
			var dst bytes.Buffer
			n, err = Copy(&dst, bytes.NewBufferString("next transfer"), buf)
			if err != nil || n != int64(len("next transfer")) || dst.String() != "next transfer" {
				t.Errorf("reused buffer Copy() = %d, %v, %q", n, err, dst.String())
			}
		})
	}
}

func TestCopyPreservesSourceError(t *testing.T) {
	srcErr := errors.New("source failed")
	src, writer := io.Pipe()
	go func() {
		writer.Write([]byte("partial source"))
		writer.CloseWithError(srcErr)
	}()
	defer src.Close()
	n, err := Copy(ioutil.Discard, src, buffer.New(4))
	if n != int64(len("partial source")) || err != srcErr {
		t.Errorf("Copy() = %d, %v; want %d, %v", n, err, len("partial source"), srcErr)
	}
}
