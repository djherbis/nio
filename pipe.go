package nio

import (
	"io"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/channel"
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

func Pipe() (io.ReadCloser, io.WriteCloser) {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	go func() {
		Copy(outWriter, inReader)
		outWriter.Close()
	}()

	return &sdReader{
		ReadCloser: outReader,
		in:         inReader,
	}, inWriter
}

func Copy(dst io.Writer, src io.Reader) (n int64, err error) {

	input := make(chan interface{})
	output := make(chan interface{})

	pending := NewBufferQueue(buffer.NewUnboundedBuffer(32*1024, 100*1024*1024))

	go channel.ChanQueue(input, output, pending)
	go inFeed(src, input)

	return outFeed(dst, output)
}

func inFeed(r io.Reader, in chan<- interface{}) {
	for {
		data := make([]byte, 32*1024)
		if n, err := r.Read(data); err == nil {
			in <- data[:n]
		} else {
			close(in)
			return
		}
	}
}

func outFeed(w io.Writer, out <-chan interface{}) (n int64, err error) {
	for output := range out {
		data := output.([]byte)
		for len(data) > 0 {
			m, err := w.Write(data)
			n += int64(m)
			data = data[m:]
			if err != nil {
				return n, err
			}
		}
	}
	return n, err
}
