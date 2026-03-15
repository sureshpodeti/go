package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================================================
// ANTI-PATTERN: Goroutine Leaks (What NOT to do)
// ============================================================================

// ✗ WRONG: Goroutine has no way to exit
func leakyWorker() {
	fmt.Println("\n" + "="*70)
	fmt.Println("ANTI-PATTERN 1: Goroutine Leak - No Exit Strategy")
	fmt.Println("="*70)

	fmt.Printf("Goroutines before: %d\n", runtime.NumGoroutine())

	// This goroutine will run forever!
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println("Working... (will never stop!)")
		}
	}()

	time.Sleep(3 * time.Second)
	fmt.Printf("Goroutines after: %d (LEAKED!)\n", runtime.NumGoroutine())
	fmt.Println("⚠️  Goroutine is still running and will never stop!")
}

// ✗ WRONG: Blocking on channel with no way to unblock
func leakyChannelWorker() {
	fmt.Println("\n" + "="*70)
	fmt.Println("ANTI-PATTERN 2: Goroutine Blocked Forever on Channel")
	fmt.Println("="*70)

	ch := make(chan int)

	fmt.Printf("Goroutines before: %d\n", runtime.NumGoroutine())

	// This goroutine will block forever waiting for data
	go func() {
		val := <-ch // Blocks forever - no one sends to this channel!
		fmt.Println("Received:", val)
	}()

	time.Sleep(2 * time.Second)
	fmt.Printf("Goroutines after: %d (LEAKED!)\n", runtime.NumGoroutine())
	fmt.Println("⚠️  Goroutine is blocked forever waiting on channel!")
}

// ============================================================================
// PATTERN 1: Done Channel (Classic Pattern)
// ============================================================================

func doneChannelPattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 1: Done Channel - Classic Pattern")
	fmt.Println("="*70)

	done := make(chan struct{}) // Signal-only channel

	fmt.Printf("Goroutines before: %d\n", runtime.NumGoroutine())

	// Worker goroutine with exit strategy
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Println("  Working...")
			case <-done:
				fmt.Println("  → Received done signal, exiting gracefully")
				return // Clean exit!
			}
		}
	}()

	time.Sleep(2 * time.Second)
	fmt.Println("\nSending done signal...")
	close(done) // Signal all listeners

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Goroutines after: %d (CLEANED UP!)\n", runtime.NumGoroutine())
	fmt.Println("✓ Goroutine exited cleanly")
}

// ============================================================================
// PATTERN 2: Context-Based Done (Modern Pattern)
// ============================================================================

func contextDonePattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 2: Context-Based Done - Modern Pattern")
	fmt.Println("="*70)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("Goroutines before: %d\n", runtime.NumGoroutine())

	// Worker goroutine using context
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Println("  Working...")
			case <-ctx.Done():
				fmt.Println("  → Context cancelled, exiting gracefully")
				return // Clean exit!
			}
		}
	}()

	time.Sleep(2 * time.Second)
	fmt.Println("\nCancelling context...")
	cancel()

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Goroutines after: %d (CLEANED UP!)\n", runtime.NumGoroutine())
	fmt.Println("✓ Goroutine exited cleanly")
}

// ============================================================================
// PATTERN 3: Multiple Workers with WaitGroup
// ============================================================================

func multipleWorkersPattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 3: Multiple Workers with WaitGroup")
	fmt.Println("="*70)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	numWorkers := 3

	fmt.Printf("Goroutines before: %d\n", runtime.NumGoroutine())
	fmt.Printf("Starting %d workers...\n", numWorkers)

	// Start multiple workers
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, i)
	}

	time.Sleep(2 * time.Second)
	fmt.Println("\nShutting down all workers...")
	cancel() // Signal all workers to stop

	wg.Wait() // Wait for all workers to finish
	fmt.Printf("Goroutines after: %d (ALL CLEANED UP!)\n", runtime.NumGoroutine())
	fmt.Println("✓ All workers exited cleanly")
}

func worker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("  Worker %d: working...\n", id)
		case <-ctx.Done():
			fmt.Printf("  → Worker %d: shutting down\n", id)
			return
		}
	}
}

// ============================================================================
// PATTERN 4: Pipeline with Done Channel
// ============================================================================

func pipelinePattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 4: Pipeline with Done Channel")
	fmt.Println("="*70)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Stage 1: Generate numbers
	numbers := generate(ctx, 1, 2, 3, 4, 5)

	// Stage 2: Square numbers
	squares := square(ctx, numbers)

	// Stage 3: Consume results
	fmt.Println("Processing pipeline...")
	count := 0
	for n := range squares {
		fmt.Printf("  Result: %d\n", n)
		count++
		if count == 3 {
			fmt.Println("\n  Cancelling pipeline early...")
			cancel() // Cancel pipeline
			break
		}
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Println("✓ Pipeline shut down cleanly (no goroutine leaks)")
}

func generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				fmt.Println("  → Generator: stopping")
				return
			}
		}
	}()
	return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				fmt.Println("  → Squarer: stopping")
				return
			}
		}
	}()
	return out
}

// ============================================================================
// PATTERN 5: Real-World HTTP Server with Graceful Shutdown
// ============================================================================

type Server struct {
	workers []chan struct{}
	wg      sync.WaitGroup
}

func (s *Server) Start(ctx context.Context) {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 5: HTTP Server with Graceful Shutdown")
	fmt.Println("="*70)

	numWorkers := 3
	s.workers = make([]chan struct{}, numWorkers)

	fmt.Printf("Starting server with %d workers...\n", numWorkers)

	// Start background workers
	for i := 0; i < numWorkers; i++ {
		done := make(chan struct{})
		s.workers[i] = done
		s.wg.Add(1)
		go s.backgroundWorker(ctx, done, i+1)
	}

	// Simulate server running
	time.Sleep(2 * time.Second)

	fmt.Println("\nReceived shutdown signal (SIGTERM)...")
	s.Shutdown()
}

func (s *Server) backgroundWorker(ctx context.Context, done chan struct{}, id int) {
	defer s.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("  Worker %d: processing requests...\n", id)
		case <-done:
			fmt.Printf("  → Worker %d: finishing current work...\n", id)
			time.Sleep(200 * time.Millisecond) // Simulate cleanup
			fmt.Printf("  → Worker %d: shutdown complete\n", id)
			return
		case <-ctx.Done():
			fmt.Printf("  → Worker %d: context cancelled\n", id)
			return
		}
	}
}

func (s *Server) Shutdown() {
	fmt.Println("\nInitiating graceful shutdown...")

	// Signal all workers to stop
	for _, done := range s.workers {
		close(done)
	}

	// Wait for all workers to finish
	s.wg.Wait()
	fmt.Println("✓ Server shutdown complete - all workers stopped cleanly")
}

// ============================================================================
// PATTERN 6: Timeout with Done Channel
// ============================================================================

func timeoutPattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 6: Timeout with Done Channel")
	fmt.Println("="*70)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := make(chan string, 1)

	// Start long-running operation
	go func() {
		// Simulate slow operation
		time.Sleep(5 * time.Second)
		result <- "Operation completed"
	}()

	fmt.Println("Starting operation with 2s timeout...")

	select {
	case res := <-result:
		fmt.Println("✓", res)
	case <-ctx.Done():
		fmt.Println("✗ Operation timed out:", ctx.Err())
		fmt.Println("✓ Goroutine will eventually finish but won't block us")
	}
}

// ============================================================================
// PATTERN 7: Fan-Out/Fan-In with Done
// ============================================================================

func fanOutFanInPattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("PATTERN 7: Fan-Out/Fan-In with Done Channel")
	fmt.Println("="*70)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Input channel
	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 1; i <= 5; i++ {
			input <- i
		}
	}()

	// Fan-out: Multiple workers process input
	numWorkers := 3
	workers := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = fanOutWorker(ctx, input, i+1)
	}

	// Fan-in: Merge results
	results := fanIn(ctx, workers...)

	fmt.Println("Processing with fan-out/fan-in...")
	count := 0
	for result := range results {
		fmt.Printf("  Result: %d\n", result)
		count++
		if count == 3 {
			fmt.Println("\n  Cancelling early...")
			cancel()
			break
		}
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Println("✓ All workers stopped cleanly")
}

func fanOutWorker(ctx context.Context, input <-chan int, id int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range input {
			select {
			case out <- n * n:
			case <-ctx.Done():
				fmt.Printf("  → Worker %d: stopping\n", id)
				return
			}
		}
	}()
	return out
}

func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
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

// ============================================================================
// BEST PRACTICES SUMMARY
// ============================================================================

func bestPracticesSummary() {
	fmt.Println("\n" + "="*70)
	fmt.Println("BEST PRACTICES SUMMARY")
	fmt.Println("="*70)
	fmt.Println(`
1. ALWAYS provide an exit strategy for goroutines
   ✓ Use context.Context or done channel
   ✗ Never start goroutines that run forever with no way to stop

2. Use select with done channel in loops
   for {
       select {
       case <-work:
           // do work
       case <-ctx.Done():
           return  // Clean exit
       }
   }

3. Use WaitGroup to wait for goroutines to finish
   var wg sync.WaitGroup
   wg.Add(1)
   go func() {
       defer wg.Done()
       // work
   }()
   wg.Wait()

4. Close channels to signal completion
   close(done)  // Signals all listeners

5. Defer cleanup in goroutines
   defer ticker.Stop()
   defer close(out)
   defer wg.Done()

6. Use buffered channels to prevent blocking
   result := make(chan string, 1)  // Won't block sender

7. Context is the modern, idiomatic way
   Prefer context.Context over plain done channels

8. Test for goroutine leaks
   runtime.NumGoroutine() before and after
`)
}

func main() {
	// Anti-patterns (what NOT to do)
	// Uncomment to see leaks (warning: will leak goroutines!)
	// leakyWorker()
	// leakyChannelWorker()

	// Correct patterns
	doneChannelPattern()
	time.Sleep(1 * time.Second)

	contextDonePattern()
	time.Sleep(1 * time.Second)

	multipleWorkersPattern()
	time.Sleep(1 * time.Second)

	pipelinePattern()
	time.Sleep(1 * time.Second)

	server := &Server{}
	ctx := context.Background()
	server.Start(ctx)
	time.Sleep(1 * time.Second)

	timeoutPattern()
	time.Sleep(1 * time.Second)

	fanOutFanInPattern()
	time.Sleep(1 * time.Second)

	bestPracticesSummary()
}
