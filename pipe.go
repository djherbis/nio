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

func Copy(dst io.Writer, src io.Reader, buf ...buffer.Buffer) (n int64, err error) {

	input := make(chan interface{})
	output := make(chan interface{})

	if len(buf) == 0 {
		buf = append(buf, buffer.NewUnboundedBuffer(32*1024, 100*1024*1024))
	}

	pending := NewBufferQueue(buffer.NewMulti(buf...))

	go channel.ChanQueue(input, output, pending)
	go inFeed(src, input)

	return outFeed(dst, output)
}

func inFeed(r io.Reader, in chan<- interface{}) {
	data := make([]byte, 32*1024)
	p := NewPacket()
	for {
		if n, err := r.Read(data); err == nil {
			p.data = data[:n]
			in <- p
			<-p.ok
		} else {
			close(in)
			return
		}
	}
}

func outFeed(w io.Writer, out <-chan interface{}) (n int64, err error) {
	data := make([]byte, 32*1024)
	for output := range out {
		p := output.(*Packet)
		x := copy(data, p.data)
		<-p.ok
		var m int
		for len(data[m:x]) > 0 {
			m, err = w.Write(data[m:x])
			n += int64(m)
			if err != nil {
				return n, err
			}
		}
	}
	return n, err
}
