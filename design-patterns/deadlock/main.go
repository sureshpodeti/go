package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/*
	Prevent Deadlocks — three core strategies:

	1. Don't create circular waits — if goroutines need multiple locks, always acquire them in the same order. No cycle, no deadlock.
	2. Don't block forever — use timeouts, select, or context so goroutines can bail out instead of waiting indefinitely.
	3. Minimize lock scope — hold locks for the shortest time possible, never hold a lock while waiting on a channel or another lock.

Do you need multiple locks?
├── No, just one lock → No deadlock possible. You're fine.
└── Yes, multiple locks needed

	├── Can you redesign to use channels instead? → Do it. Single owner, no lock ordering problem.
	└── No, must use multiple locks
	    ├── Always acquire locks in the same order → Breaks circular wait.
	    └── Can't guarantee order?
	        ├── Use TryLock → Skip if busy, try again later.
	        └── Use timeouts (select + time.After / context) → Bail out instead of waiting forever.

Are you using channels?
├── Unbuffered and sender/receiver might not be ready at the same time?
│   ├── Use buffered channels → Decouples sender from receiver in time.
│   └── Use select with timeout → Don't block forever on send or receive.
└── Circular channel dependency (A waits on B, B waits on A)?

	└── Redesign the flow → Break the cycle, introduce a coordinator goroutine.

General safety net for all cases:
├── Hold locks for the shortest time possible → Lock, do minimal work, unlock.
├── Never hold a lock while waiting on a channel or another lock.
└── Always use defer mu.Unlock() → Guarantees release even on panic.

	Summary: Mental model

Can the goroutine make progress without waiting on anyone?

├── Yes → No deadlock risk. Done.
└── No, it waits on something

	├── Does what it's waiting on also wait on it? (circular)
	│   → Break the cycle — reorder locks, redesign channel flow,
	│     or add a coordinator.
	└── No cycle, but it might wait forever
	    → Add timeouts (select + time.After / context.WithTimeout).
*/
func CreateDeadlock() {
	ch1, ch2 := make(chan struct{}), make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		<-ch2
		ch1 <- struct{}{}

	}()

	go func() {
		wg.Done()
		<-ch1
		ch2 <- struct{}{}
	}()

	wg.Wait()
}

func DeadlockFixConsistentOrdering() {
	ch1, ch2 := make(chan struct{}), make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		<-ch1
		ch2 <- struct{}{}
	}()

	go func() {
		wg.Done()
		ch1 <- struct{}{}
		<-ch2
	}()

	wg.Wait()
}

func DeadlockFixWithTimeout() {
	ch1, ch2 := make(chan struct{}), make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()

		select {
		case <-ch1:
			ch2 <- struct{}{}
			fmt.Println("Goroutine 1 is completed!")
		case <-time.After(2 * time.Second):
			fmt.Println("Goroutine 1 timedout after 2 seconds")
		}

	}()

	go func() {
		defer wg.Done()
		select {
		case <-ch2:
			ch1 <- struct{}{}
			fmt.Println("Goroutine 2 is completed!")

		case <-time.After(2 * time.Second):
			fmt.Println("Goroutine 2 timeout after 2 seconds")
		}
	}()

	wg.Wait()
}

func DeadlockFixWithContext() {
	ch1, ch2 := make(chan struct{}), make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()

		select {
		case <-ch1:
			ch2 <- struct{}{}
			fmt.Println("Goroutine 1 completed!")
		case <-ctx.Done():
			fmt.Println("Goroutine 1 error : ", ctx.Err())
		}

	}()

	go func() {
		defer wg.Done()
		select {
		case <-ch2:
			ch1 <- struct{}{}
			fmt.Println("Goroutine 2 completed!")
		case <-ctx.Done():
			fmt.Println("Goroutine 2 error : ", ctx.Err())
		}

	}()

	wg.Wait()

}

func DeadlockFixWithBufferedChannels() {
	ch1, ch2 := make(chan struct{}, 1), make(chan struct{}, 1)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		ch1 <- struct{}{}
		<-ch2
	}()

	go func() {
		defer wg.Done()

		ch2 <- struct{}{}
		<-ch1
	}()

	wg.Wait()
}
func main() {
	// CreateDeadlock()
	// DeadlockFixConsistentOrdering()
	// DeadlockFixWithTimeout()
	// DeadlockFixWithContext()
	DeadlockFixWithBufferedChannels()
}
