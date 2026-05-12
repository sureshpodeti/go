package main

import (
	"fmt"
	"time"
)

func counter(input <-chan string, done chan<- struct{}) {

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-input:
			fmt.Println("Got input")
			close(done)
			return
		case <-tick.C:
			fmt.Println("TICK")

		}
	}
}
func main() {
	done := make(chan struct{})
	input := make(chan string)

	go counter(input, done)
	go func() {
		defer close(input)
		time.Sleep(time.Second * 10)
		input <- "hello"
	}()

	<-done
}
