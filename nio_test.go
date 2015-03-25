package nio

import (
	"bytes"
	"testing"

	"github.com/djherbis/buffer"
)

func TestCopy(t *testing.T) {
	buf := buffer.New(1024)
	input := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)

	input.Write([]byte("hello world"))
	n, err := Copy(output, input, buf)
	if err != nil {
		t.Errorf(err.Error())
	}
	if int(n) != len("hello world") {
		t.Errorf("wrote wrong # of bytes")
	}

	if !bytes.Equal(output.Bytes(), []byte("hello world")) {
		t.Errorf("output didn't match: %s", output.Bytes())
	}
}

func TestPipe(t *testing.T) {
	buf := buffer.New(1024)
	r, w := Pipe(buf)

	data := []byte("the quick brown fox jumps over the lazy dog")
	if _, err := w.Write(data); err != nil {
		t.Error(err.Error())
		return
	}

	result := make([]byte, 1024)
	n, err := r.Read(result)
	if err != nil {
		t.Error(err.Error())
		return
	}
	result = result[:n]

	if !bytes.Equal(data, result) {
		t.Errorf("exp [%s]\ngot[%s]", string(data), string(result))
	}

}
