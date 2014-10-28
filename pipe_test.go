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

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(data, []byte("0123456789")) {
		t.Error("Not equal!")
	}
}
