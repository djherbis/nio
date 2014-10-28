package nio

import "io"

func NewWriter(writer io.Writer) io.WriteCloser {
	r, w := Pipe()
	go io.Copy(writer, r)
	return w
}

func NewWriteCloser(writer io.WriteCloser) io.WriteCloser {
	r, w := Pipe()
	go func() {
		io.Copy(writer, r)
		writer.Close()
	}()
	return w
}
