package nio

import (
	"io"
	"lib/buffer"
	"lib/channel"
)

func Pipe() (io.ReadCloser, io.WriteCloser) {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	input := make(chan interface{})
	output := make(chan interface{})

	pending := buffer.NewBufferQueue(buffer.NewUnboundedBuffer(32*1024, 100*1024*1024))
	go channel.ChanQueue(input, output, pending)
	go inFeed(inReader, input)
	go outFeed(outWriter, output)

	return outReader, inWriter
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
			n, _ := w.Write(data)
			data = data[n:]
		}

	}
	w.Close()
}
