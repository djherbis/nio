package channel

import (
	"fmt"
	"sync"
	"testing"
)

func TestMemChan(t *testing.T) {
	input := make(chan interface{})
	output := make(chan interface{})
	go Chan(input, output)

	group := new(sync.WaitGroup)
	group.Add(1)

	go func() {
		for i := 0; i < 10; i++ {
			input <- []byte(fmt.Sprintf("%d", i))
		}

		go func() {
			defer group.Done()
			s := ""
			for out := range output {
				s += string(out.([]byte))
			}
			if s != "0123456789" {
				t.Error("Not equal!")
			}
		}()

		close(input)

	}()

	group.Wait()
}

func TestChan(t *testing.T) {
	input := make(chan interface{})
	output := make(chan interface{})
	go Chan(input, output)

	group := new(sync.WaitGroup)
	group.Add(1)

	go func() {
		for i := 0; i < 10; i++ {
			input <- []byte(fmt.Sprintf("%d", i))
		}

		go func() {
			defer group.Done()
			s := ""
			for out := range output {
				s += string(out.([]byte))
			}
			if s != "0123456789" {
				t.Error("Not equal!")
			}
		}()

		close(input)

	}()

	group.Wait()
}
