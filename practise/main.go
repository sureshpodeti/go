package main

import (
	"context"
	"fmt"
	"sync"
)

// Stage 1 (Pipeline): Generate numbers 1-100
func generate(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= 100; i++ {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage 2 (Fan-Out): Split into even and odd channels
func split(ctx context.Context, in <-chan int) (<-chan int, <-chan int) {
	even := make(chan int)
	odd := make(chan int)
	go func() {
		defer close(even)
		defer close(odd)
		for n := range in {
			if n%2 == 0 {
				select {
				case even <- n:
				case <-ctx.Done():
					return
				}
			} else {
				select {
				case odd <- n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return even, odd
}

// Stage 3 (Fan-In): Merge even and odd back into one channel
func merge(ctx context.Context, channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for n := range c {
				select {
				case out <- n:
				case <-ctx.Done():
					return
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Stage 4 (Pipeline): Print results
func print(ctx context.Context, in <-chan int) {
	for n := range in {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println(n)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Pipeline:  generate → split → merge → print
	//                        ├─ even ─┐
	//                        └─ odd  ─┘

	numbers := generate(ctx)
	even, odd := split(ctx, numbers)
	merged := merge(ctx, even, odd)
	print(ctx, merged)
}
