package nio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPipe(t *testing.T) {
	r, w := Pipe()

	go func() {
		for i := 0; i < 10; i++ {
			w.Write([]byte(fmt.Sprintf("%d", i)))
		}
		w.Close()
	}()

	data, _ := ioutil.ReadAll(r)
	if !bytes.Equal(data, []byte("0123456789")) {
		t.Error("Not equal!")
	}
}
