package nio

import (
	"bytes"
	"io"
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

func TestBigWriteSmallBuf(t *testing.T) {
	buf := buffer.New(5)
	r, w := Pipe(buf)
	defer r.Close()

	go func() {
		defer w.Close()
		n, err := w.Write([]byte("hello world"))
		if err != nil {
			t.Error(err.Error())
		}
		if int(n) != len("hello world") {
			t.Errorf("wrote wrong # of bytes")
		}
	}()

	data := make([]byte, 24)
	n, err := r.Read(data)
	if err != nil && err != io.EOF {
		t.Error(err.Error())
	}
	data = data[:n]
	if !bytes.Equal(data, []byte("hello world")) {
		t.Errorf("unexpected output %s", data)
	}
}

func TestPipeCloseEarly(t *testing.T) {
	buf := buffer.New(1024)
	r, w := Pipe(buf)
	r.Close()
	_, err := w.Write([]byte("hello world"))
	if err != io.ErrClosedPipe {
		t.Errorf("expected closed pipe")
	}
}

func TestPipe(t *testing.T) {
	buf := buffer.New(1024)
	r, w := Pipe(buf)
	defer r.Close()

	data := []byte("the quick brown fox jumps over the lazy dog")
	if _, err := w.Write(data); err != nil {
		t.Error(err.Error())
		return
	}
	w.Close()

	result := make([]byte, 1024)
	n, err := r.Read(result)
	if err != nil && err != io.EOF {
		t.Error(err.Error())
		return
	}
	result = result[:n]

	if !bytes.Equal(data, result) {
		t.Errorf("exp [%s]\ngot[%s]", string(data), string(result))
	}

}
