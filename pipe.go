package nio

import (
	"io"
	"lib/buffer"
	"lib/channel"
)

type sdReader struct {
	io.ReadCloser
	out   io.ReadCloser
	chout chan interface{}
}

func (r *sdReader) Close() error {
	r.ReadCloser.Close()
	r.out.Close()
	for _ = range r.chout {
	}
	return nil
}

func Pipe() (io.ReadCloser, io.WriteCloser) {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	input := make(chan interface{})
	output := make(chan interface{})

	pending := buffer.NewBufferQueue(buffer.NewFile(1024))
	go channel.ChanQueue(input, output, pending)
	go inFeed(inReader, input)
	go outFeed(outWriter, output)

	return &sdReader{
		ReadCloser: inReader,
		out:        outReader,
		chout:      output,
	}, inWriter
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

func outFeed(w io.WriteCloser, out <-chan interface{}) {
	for output := range out {
		data := output.([]byte)
		for len(data) > 0 {
			if n, err := w.Write(data); err == nil {
				data = data[n:]
			} else {
				return
			}
		}

	}
	w.Close()
}
