package queue

type Queue interface {
	Peek() interface{}
	Pop() interface{}
	Push(interface{})
	Len() int64
}

func Empty(q Queue) bool {
	return q.Len() == 0
}

type SliceQueue []interface{}

func NewSliceQueue() Queue {
	q := SliceQueue(make([]interface{}, 0))
	return &q
}

func (q *SliceQueue) Push(b interface{}) {
	*q = append(*q, b)
}

func (q *SliceQueue) Peek() (b interface{}) {
	if len(*q) > 0 {
		b = (*q)[0]
	}
	return b
}

func (q *SliceQueue) Pop() (b interface{}) {
	if len(*q) > 0 {
		b = (*q)[0]
		*q = (*q)[1:]
	}
	return b
}

func (q *SliceQueue) Len() int64 {
	return int64(len(*q))
}
