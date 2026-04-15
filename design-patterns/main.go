package main

import (
	"fmt"
	"sort"
)

// Lift simulates the LOOK disk/elevator scheduling algorithm.
// It starts at the given position, services requests in the current
// direction, then reverses when there are no more requests ahead.
func Lift(requests []int, start int) []int {
	if len(requests) == 0 {
		return nil
	}

	sorted := make([]int, len(requests))
	copy(sorted, requests)
	sort.Ints(sorted)

	// Split into requests below and at/above the start position
	var lower, upper []int
	for _, r := range sorted {
		if r < start {
			lower = append(lower, r)
		} else {
			upper = append(upper, r)
		}
	}

	// Service upper first (going up), then reverse through lower
	order := make([]int, 0, len(requests))
	order = append(order, upper...)
	// Reverse lower so we service highest-first on the way down
	for i := len(lower) - 1; i >= 0; i-- {
		order = append(order, lower[i])
	}

	return order
}

func totalMovement(order []int, start int) int {
	total := 0
	current := start
	for _, pos := range order {
		diff := current - pos
		if diff < 0 {
			diff = -diff
		}
		total += diff
		current = pos
	}
	return total
}

func main() {
	requests := []int{176, 79, 34, 60, 92, 11, 41, 114}
	start := 50

	order := Lift(requests, start)

	fmt.Println("Start position:", start)
	fmt.Println("Request queue:", requests)
	fmt.Println("Service order:", order)
	fmt.Println("Total movement:", totalMovement(order, start))
}
