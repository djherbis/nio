package nio

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func BenchmarkIOCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.Copy(ioutil.Discard, io.LimitReader(rand.Reader, 10*1024*1024))
	}
}

func BenchmarkNioCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Copy(ioutil.Discard, io.LimitReader(rand.Reader, 10*1024*1024))
	}
}

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
		t.Error("Not equal! " + string(data))
	}
}
