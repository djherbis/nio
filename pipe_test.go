package nio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPipe(t *testing.T) {
	r, w := Pipe()

	r = NewReadCloser(r)
	w = NewWriteCloser(w)

	go func() {
		for i := 0; i < 10; i++ {
			if _, err := w.Write([]byte(fmt.Sprintf("%d", i))); err != nil {
				t.Error(err.Error())
				return
			}
		}
		w.Close()
	}()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(data, []byte("0123456789")) {
		t.Error("Not equal!")
	}
}
