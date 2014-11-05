package nio

import "github.com/djherbis/buffer"

var empty struct{}

type Packet struct {
	data []byte
	ok   chan struct{}
}

func NewPacket() *Packet {
	return &Packet{
		ok: make(chan struct{}),
	}
}

type BufferQueue struct {
	top  *Packet
	data []byte
	buffer.Buffer
}

func NewBufferQueue(buf buffer.Buffer) *BufferQueue {
	return &BufferQueue{
		top:    NewPacket(),
		data:   make([]byte, 32*1024),
		Buffer: buf,
	}
}

func (q *BufferQueue) Len() (n int64) {
	return q.Buffer.Len()
}

func (q *BufferQueue) Peek() (b interface{}) {
	n, _ := q.Buffer.ReadAt(q.data, 0)
	q.top.data = q.data[:n]
	return q.top
}

func (q *BufferQueue) Pop() (b interface{}) {
	q.top.ok <- empty
	q.Buffer.FastForward(len(q.top.data))
	return q.top
}

func (q *BufferQueue) Push(b interface{}) {
	p := b.(*Packet)
	q.Write(p.data)
	p.ok <- empty
}
