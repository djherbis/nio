package nio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/djherbis/buffer"
)

func TestPipe(t *testing.T) {
	r, w := Pipe(buffer.New(32 * 1024))

	r = NewReadCloser(r, buffer.New(32*1024))
	w = NewWriteCloser(w, buffer.New(32*1024))

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
		t.Error("Not equal! " + string(data))
	}
}
