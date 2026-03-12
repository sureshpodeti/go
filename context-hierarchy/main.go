package main

import (
	"context"
	"fmt"
	"time"
)

// Simulates a database query
func queryDatabase(ctx context.Context, query string) error {
	select {
	case <-time.After(3 * time.Second):
		fmt.Println("✓ Database query completed:", query)
		return nil
	case <-ctx.Done():
		fmt.Println("✗ Database query cancelled:", query, "- Reason:", ctx.Err())
		return ctx.Err()
	}
}

// Simulates calling an external API
func callExternalAPI(ctx context.Context, endpoint string) error {
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("✓ API call completed:", endpoint)
		return nil
	case <-ctx.Done():
		fmt.Println("✗ API call cancelled:", endpoint, "- Reason:", ctx.Err())
		return ctx.Err()
	}
}

// Simulates processing cache
func checkCache(ctx context.Context, key string) error {
	select {
	case <-time.After(1 * time.Second):
		fmt.Println("✓ Cache check completed:", key)
		return nil
	case <-ctx.Done():
		fmt.Println("✗ Cache check cancelled:", key, "- Reason:", ctx.Err())
		return ctx.Err()
	}
}

// Child operation with its own shorter timeout
func fetchUserProfile(ctx context.Context, userID string) error {
	// Child context: 2 second timeout (shorter than parent's 5 seconds)
	childCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	fmt.Println("\n→ Starting fetchUserProfile (2s timeout)")

	// This will timeout after 2 seconds even though parent has 5 seconds
	err := queryDatabase(childCtx, "SELECT * FROM users WHERE id="+userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user profile: %w", err)
	}

	return nil
}

// Another child operation with different timeout
func fetchUserOrders(ctx context.Context, userID string) error {
	// Child context: 4 second timeout
	childCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	fmt.Println("\n→ Starting fetchUserOrders (4s timeout)")

	err := queryDatabase(childCtx, "SELECT * FROM orders WHERE user_id="+userID)
	if err != nil {
		return fmt.Errorf("failed to fetch orders: %w", err)
	}

	return nil
}

// Parent operation that orchestrates multiple child operations
func handleUserRequest(ctx context.Context, userID string) error {
	fmt.Println("\n=== Handling User Request ===")

	// Step 1: Check cache (uses parent context)
	if err := checkCache(ctx, "user:"+userID); err != nil {
		return err
	}

	// Step 2: Fetch user profile (has its own 2s timeout)
	if err := fetchUserProfile(ctx, userID); err != nil {
		return err
	}

	// Step 3: Fetch orders (has its own 4s timeout)
	if err := fetchUserOrders(ctx, userID); err != nil {
		return err
	}

	// Step 4: Call external API (uses parent context)
	if err := callExternalAPI(ctx, "/api/user-analytics"); err != nil {
		return err
	}

	fmt.Println("\n✓ User request completed successfully")
	return nil
}

// Demonstrates parent cancellation cascading to children
func demonstrateParentCancellation() {
	fmt.Println("DEMO 1: Parent Cancellation Cascades to All Children")

	// Parent context with 5 second timeout
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer parentCancel()

	// Create multiple child contexts
	child1Ctx, cancel1 := context.WithTimeout(parentCtx, 10*time.Second) // Wants 10s but parent only gives 5s
	defer cancel1()

	child2Ctx, cancel2 := context.WithTimeout(parentCtx, 8*time.Second) // Wants 8s but parent only gives 5s
	defer cancel2()

	fmt.Println("Parent timeout: 5s")
	fmt.Println("Child1 timeout: 10s (but will be cancelled by parent at 5s)")
	fmt.Println("Child2 timeout: 8s (but will be cancelled by parent at 5s)")

	// Start operations with child contexts
	go func() {
		<-child1Ctx.Done()
		fmt.Println("\n→ Child1 cancelled:", child1Ctx.Err())
	}()

	go func() {
		<-child2Ctx.Done()
		fmt.Println("→ Child2 cancelled:", child2Ctx.Err())
	}()

	// Wait for parent to timeout
	<-parentCtx.Done()
	fmt.Println("→ Parent cancelled:", parentCtx.Err())

	time.Sleep(100 * time.Millisecond) // Give goroutines time to print
}

// Demonstrates child can have shorter timeout than parent
func demonstrateChildShorterTimeout() {
	fmt.Println("DEMO 2: Child Has Shorter Timeout Than Parent")

	// Parent context with 10 second timeout
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer parentCancel()

	// Child context with 2 second timeout (shorter than parent)
	childCtx, childCancel := context.WithTimeout(parentCtx, 2*time.Second)
	defer childCancel()

	fmt.Println("Parent timeout: 10s")
	fmt.Println("Child timeout: 2s")

	// Child will timeout first
	select {
	case <-childCtx.Done():
		fmt.Println("\n→ Child timed out first:", childCtx.Err())
		fmt.Println("→ Parent is still active:", parentCtx.Err()) // Will be nil
	}
}

// Real-world example: HTTP request handler with nested operations
func demonstrateRealWorldScenario() {
	fmt.Println("DEMO 3: Real-World HTTP Request Handler")

	// Simulate incoming HTTP request with 10 second timeout
	requestCtx, requestCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer requestCancel()

	fmt.Println("HTTP Request timeout: 10s")

	// Handle the request
	if err := handleUserRequest(requestCtx, "user123"); err != nil {
		fmt.Println("\n✗ Request failed:", err)
	}
}

// Demonstrates manual parent cancellation
func demonstrateManualCancellation() {
	fmt.Println("DEMO 4: Manual Parent Cancellation")

	// Parent context with manual cancellation
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Create child contexts
	child1Ctx, cancel1 := context.WithCancel(parentCtx)
	defer cancel1()

	child2Ctx, cancel2 := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel2()

	fmt.Println("Parent: manual cancellation")
	fmt.Println("Child1: manual cancellation")
	fmt.Println("Child2: 10s timeout")

	// Start operations
	go func() {
		<-child1Ctx.Done()
		fmt.Println("\n→ Child1 cancelled:", child1Ctx.Err())
	}()

	go func() {
		<-child2Ctx.Done()
		fmt.Println("→ Child2 cancelled:", child2Ctx.Err())
	}()

	// Manually cancel parent after 2 seconds
	time.Sleep(2 * time.Second)
	fmt.Println("\n→ Manually cancelling parent...")
	parentCancel()

	time.Sleep(100 * time.Millisecond) // Give goroutines time to print
	fmt.Println("→ All children were cancelled when parent was cancelled")
}

// Demonstrates context tree structure
func demonstrateContextTree() {
	fmt.Println("DEMO 5: Context Tree Structure")

	// Root
	root := context.Background()

	// Level 1: Request context
	requestCtx, requestCancel := context.WithTimeout(root, 10*time.Second)
	defer requestCancel()

	// Level 2: Service layer contexts
	authCtx, authCancel := context.WithTimeout(requestCtx, 3*time.Second)
	defer authCancel()

	dataCtx, dataCancel := context.WithTimeout(requestCtx, 5*time.Second)
	defer dataCancel()

	// Level 3: Individual operation contexts
	dbCtx, dbCancel := context.WithTimeout(dataCtx, 2*time.Second)
	defer dbCancel()

	cacheCtx, cacheCancel := context.WithTimeout(dataCtx, 1*time.Second)
	defer cacheCancel()

	fmt.Println("Context Tree:")
	fmt.Println("  Background (root)")
	fmt.Println("    └─ RequestCtx (10s)")
	fmt.Println("         ├─ AuthCtx (3s)")
	fmt.Println("         └─ DataCtx (5s)")
	fmt.Println("              ├─ DbCtx (2s)")
	fmt.Println("              └─ CacheCtx (1s)")

	fmt.Println("\nCancelling RequestCtx will cancel all children...")
	requestCancel()

	// Check all contexts
	time.Sleep(100 * time.Millisecond)
	fmt.Println("\nContext states after parent cancellation:")
	fmt.Println("  AuthCtx:", authCtx.Err())
	fmt.Println("  DataCtx:", dataCtx.Err())
	fmt.Println("  DbCtx:", dbCtx.Err())
	fmt.Println("  CacheCtx:", cacheCtx.Err())
}

func main() {
	// demonstrateParentCancellation()
	// time.Sleep(1 * time.Second)

	// demonstrateChildShorterTimeout()
	// time.Sleep(1 * time.Second)

	// demonstrateRealWorldScenario()
	// time.Sleep(1 * time.Second)

	// demonstrateManualCancellation()
	// time.Sleep(1 * time.Second)

	demonstrateContextTree()
}
