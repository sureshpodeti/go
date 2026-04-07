package main

import (
	"fmt"
	"sync"
)

/*
	Prevent Race conditions -
		1. Don’t share the resource - each go routine works on its own data; there is nothing to race over;
		2. Transfer ownership instead of sharing  -  Use channels to transfer data from one go routine to another. Once you send it, you use stopping it. Only one go routine owns the data no conflict
		3.  If you must share - use mutex to ensure only one go routine can access the shared data

	Can each goroutine work on its own data?
├── Yes → Confinement. Done. Simplest solution.
└── No, goroutines need to exchange data
    ├── Data flows from one to another → Channels
    └── Multiple goroutines access the same thing → Mutex

*/

func CreateRaceCondition() {

	counter := 0

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}

	wg.Wait()
	fmt.Println("Counter - ", counter)
}

func RaceConditionFixConfinement() {
	ar := make([]int, 1000)

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ar[i]++
		}(i)
	}
	wg.Wait()

	counter := 0

	for i := 0; i < 1000; i++ {
		counter += ar[i]
	}

	fmt.Println(counter)
}

func RaceConditionFixWithChannels() {
	var wg sync.WaitGroup

	counter := 0

	counterCh := make(chan int)
	done := make(chan struct{})

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counterCh <- 1
		}()
	}

	go func() {
		for cc := range counterCh {
			counter += cc
		}
		close(done)
	}()

	wg.Wait()
	close(counterCh)
	<-done

	fmt.Println("Counter - ", counter)

}

func RaceConditionFixWithMutex() {

	counter := 0

	var mtx sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer mtx.Unlock()
			mtx.Lock()
			counter++
		}()
	}

	wg.Wait()
	fmt.Println("counter - ", counter)

}
func main() {
	// CreateRaceCondition()
	// RaceConditionFixConfinement()
	// RaceConditionFixWithChannels()
	RaceConditionFixWithMutex()
}
