package channel

import "github.com/djherbis/nio/queue"

func Chan(in <-chan interface{}, next chan<- interface{}) {
	ChanQueue(in, next, queue.NewSliceQueue())
}

func ChanQueue(in <-chan interface{}, next chan<- interface{}, pending queue.Queue) {
	defer close(next)

recv:

	for {

		if queue.Empty(pending) {
			data, ok := <-in
			if !ok {
				break
			}

			pending.Push(data)
		}

		select {
		case data, ok := <-in:
			if !ok {
				break recv
			}
			pending.Push(data)

		case next <- pending.Peek():
			pending.Pop()
		}

	}

	for !queue.Empty(pending) {
		next <- pending.Pop()
	}
}
