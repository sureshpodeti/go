package main

import (
	"context"
	"fmt"
	"time"
)

// mergeContexts combines multiple contexts - cancels when ANY of them cancels
func mergeContexts(ctxs ...context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Monitor all input contexts
	for _, c := range ctxs {
		go func(inputCtx context.Context) {
			select {
			case <-inputCtx.Done():
				cancel() // Cancel merged context when any input cancels
			case <-ctx.Done():
				return // Merged context already cancelled
			}
		}(c)
	}

	return ctx, cancel
}

// DEMO 1: User cancellation OR timeout (whichever comes first)
func demonstrateUserCancellationOrTimeout() {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 1: User Cancellation OR Timeout (First-to-Cancel Wins)")
	fmt.Println("="*70)

	// Context 1: User can cancel manually
	userCtx, userCancel := context.WithCancel(context.Background())

	// Context 2: 10 second timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutCancel()

	// Merge: cancel if EITHER user cancels OR timeout occurs
	mergedCtx, mergedCancel := mergeContexts(userCtx, timeoutCtx)
	defer mergedCancel()

	fmt.Println("Starting long operation...")
	fmt.Println("- Will timeout in 10 seconds")
	fmt.Println("- User will cancel in 3 seconds")

	// Simulate user cancelling after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("\n→ User clicked 'Cancel' button!")
		userCancel()
	}()

	// Long-running operation
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("✓ Operation completed")
	case <-mergedCtx.Done():
		fmt.Println("✗ Operation cancelled:", mergedCtx.Err())
		fmt.Println("  (User cancellation won the race)")
	}
}

// DEMO 2: Multiple API calls - cancel all if any fails
func demonstrateMultipleAPICalls() {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 2: Multiple API Calls - Cancel All If Any Times Out")
	fmt.Println("="*70)

	// Each API has its own timeout
	api1Ctx, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()

	api2Ctx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	api3Ctx, cancel3 := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel3()

	// Merge: if ANY API times out, cancel all operations
	mergedCtx, mergedCancel := mergeContexts(api1Ctx, api2Ctx, api3Ctx)
	defer mergedCancel()

	fmt.Println("Calling 3 APIs in parallel:")
	fmt.Println("- API 1: 5s timeout")
	fmt.Println("- API 2: 3s timeout (will timeout first)")
	fmt.Println("- API 3: 7s timeout")

	// Simulate API calls
	done := make(chan string, 3)

	// API 1
	go func() {
		select {
		case <-time.After(6 * time.Second):
			done <- "API 1 completed"
		case <-mergedCtx.Done():
			done <- "API 1 cancelled"
		}
	}()

	// API 2
	go func() {
		select {
		case <-time.After(4 * time.Second):
			done <- "API 2 completed"
		case <-mergedCtx.Done():
			done <- "API 2 cancelled"
		}
	}()

	// API 3
	go func() {
		select {
		case <-time.After(8 * time.Second):
			done <- "API 3 completed"
		case <-mergedCtx.Done():
			done <- "API 3 cancelled"
		}
	}()

	// Wait for results
	for i := 0; i < 3; i++ {
		result := <-done
		fmt.Println("→", result)
	}

	fmt.Println("\n✗ All operations cancelled because API 2 timed out first")
}

// DEMO 3: Request timeout OR server shutdown (real-world scenario)
func demonstrateRequestOrShutdown() {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 3: Request Timeout OR Server Shutdown")
	fmt.Println("="*70)

	// Context 1: Server shutdown signal
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	// Context 2: Individual request timeout
	requestCtx, requestCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer requestCancel()

	// Merge: cancel if EITHER request times out OR server shuts down
	mergedCtx, mergedCancel := mergeContexts(shutdownCtx, requestCtx)
	defer mergedCancel()

	fmt.Println("Processing request...")
	fmt.Println("- Request timeout: 10 seconds")
	fmt.Println("- Server will shutdown in 2 seconds")

	// Simulate server shutdown after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n→ Server received SIGTERM (shutdown signal)!")
		shutdownCancel()
	}()

	// Process request
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("✓ Request processed successfully")
	case <-mergedCtx.Done():
		fmt.Println("✗ Request cancelled:", mergedCtx.Err())
		fmt.Println("  (Server shutdown won the race)")
	}
}

// DEMO 4: Circuit breaker pattern
func demonstrateCircuitBreaker() {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 4: Circuit Breaker Pattern")
	fmt.Println("="*70)

	// Context 1: Normal request timeout
	requestCtx, requestCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer requestCancel()

	// Context 2: Circuit breaker trips after 3 seconds
	circuitBreakerCtx, breakerCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer breakerCancel()

	// Merge: cancel if EITHER request times out OR circuit breaker trips
	mergedCtx, mergedCancel := mergeContexts(requestCtx, circuitBreakerCtx)
	defer mergedCancel()

	fmt.Println("Making request to flaky service...")
	fmt.Println("- Request timeout: 10 seconds")
	fmt.Println("- Circuit breaker: 3 seconds")

	// Simulate slow service
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("✓ Service responded")
	case <-mergedCtx.Done():
		fmt.Println("✗ Request failed:", mergedCtx.Err())
		fmt.Println("  (Circuit breaker tripped - protecting system)")
	}
}

// DEMO 5: Parent request + child operation timeouts
func demonstrateNestedTimeouts() {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 5: Nested Timeouts - Parent Request + Child Operations")
	fmt.Println("="*70)

	// Parent: Overall request timeout (10 seconds)
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer parentCancel()

	// Child 1: Database query timeout (2 seconds)
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dbCancel()

	// Child 2: Cache timeout (1 second)
	cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cacheCancel()

	// Merge: cancel if parent times out OR any child operation times out
	mergedCtx, mergedCancel := mergeContexts(parentCtx, dbCtx, cacheCtx)
	defer mergedCancel()

	fmt.Println("Processing request with multiple operations:")
	fmt.Println("- Parent timeout: 10 seconds")
	fmt.Println("- DB query timeout: 2 seconds")
	fmt.Println("- Cache timeout: 1 second (will timeout first)")

	// Simulate operations
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("✓ All operations completed")
	case <-mergedCtx.Done():
		fmt.Println("✗ Operations cancelled:", mergedCtx.Err())
		fmt.Println("  (Cache timeout won - fastest timeout)")
	}
}

// DEMO 6: Real-world HTTP handler with multiple cancellation sources
type Server struct {
	shutdownCtx context.Context
}

func (s *Server) HandleRequest(w interface{}, r interface{}) {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 6: Real-World HTTP Handler")
	fmt.Println("="*70)

	// Context 1: Client disconnects
	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()

	// Context 2: Request timeout (30 seconds)
	requestCtx, requestCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer requestCancel()

	// Context 3: Server shutdown
	// (in real code, this would be s.shutdownCtx from graceful shutdown)

	// Merge all cancellation sources
	mergedCtx, mergedCancel := mergeContexts(clientCtx, requestCtx, s.shutdownCtx)
	defer mergedCancel()

	fmt.Println("Request started with multiple cancellation sources:")
	fmt.Println("- Client can disconnect")
	fmt.Println("- Request timeout: 30 seconds")
	fmt.Println("- Server can shutdown")
	fmt.Println("\nSimulating client disconnect in 2 seconds...")

	// Simulate client disconnect
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n→ Client disconnected!")
		clientCancel()
	}()

	// Process request
	if err := s.processRequest(mergedCtx); err != nil {
		fmt.Println("✗ Request failed:", err)
	}
}

func (s *Server) processRequest(ctx context.Context) error {
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("✓ Request processed")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("request cancelled: %w", ctx.Err())
	}
}

// DEMO 7: Using golang.org/x/sync/errgroup pattern (alternative approach)
func demonstrateErrGroupPattern() {
	fmt.Println("\n" + "="*70)
	fmt.Println("DEMO 7: Alternative - Using errgroup (Standard Pattern)")
	fmt.Println("="*70)
	fmt.Println("Note: This is the idiomatic Go way for parallel operations")
	fmt.Println("(Requires: go get golang.org/x/sync/errgroup)")

	fmt.Println(`
Example code:
	
	import "golang.org/x/sync/errgroup"
	
	g, ctx := errgroup.WithContext(context.Background())
	
	// If ANY goroutine returns error, ctx is cancelled
	g.Go(func() error {
		return fetchAPI1(ctx)
	})
	
	g.Go(func() error {
		return fetchAPI2(ctx)
	})
	
	// Wait for all, cancel all if any fails
	if err := g.Wait(); err != nil {
		// First error is returned, all others cancelled
	}
`)
}

func main() {
	demonstrateUserCancellationOrTimeout()
	time.Sleep(1 * time.Second)

	demonstrateMultipleAPICalls()
	time.Sleep(1 * time.Second)

	demonstrateRequestOrShutdown()
	time.Sleep(1 * time.Second)

	demonstrateCircuitBreaker()
	time.Sleep(1 * time.Second)

	demonstrateNestedTimeouts()
	time.Sleep(1 * time.Second)

	// Demo 6
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()
	server := &Server{shutdownCtx: shutdownCtx}
	server.HandleRequest(nil, nil)
	time.Sleep(1 * time.Second)

	demonstrateErrGroupPattern()
}
