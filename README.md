nio
==========
Documentation: http://godoc.org/github.com/djherbis/nio

Usage
-----

The Buffer interface:

```go
type Buffer interface {
	Len() int64
	Cap() int64
	io.ReadWriter
}

```

nio's Copy method concurrently copies from an io.Reader to a supplied nio.Buffer, 
then from the nio.Buffer to an io.Writer. This way, blocking writes don't slow the io.Reader.

```go
import (
  "github.com/djherbis/buffer"
  "github.com/djherbis/nio"
)

buf := buffer.New(32*1024) // 32KB In memory Buffer
nio.Copy(w, r, buf) // Reads and Writes concurrently, buffering using buf.
```

nio's Pipe method is a buffered version of io.Pipe
```go
import (
  "github.com/djherbis/buffer"
  "github.com/djherbis/nio"
)

buf := buffer.New(32*1024) // 32KB In memory Buffer
r, w := nio.Pipe(buf)
The writer return once its data has been written to the Buffer.
The reader returns with data off the Buffer.
```

Installation
------------
```sh
go get github.com/djherbis/nio
```

For some pre-built buffers grab:
```sh
go get github.com/djherbis/buffer
```
