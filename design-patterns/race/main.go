package main

import (
	"fmt"
	"sync"
)

func main() {

	var wg sync.WaitGroup

	balance := 1000
	wg.Add(2)

	// Withdraw
	go func() {
		defer wg.Done()
		current := balance
		newbalance := current - 500
		balance = newbalance
	}()

	// Deposit
	go func() {
		defer wg.Done()
		current := balance
		newbalance := current + 200
		balance = newbalance
	}()

	wg.Wait()

	fmt.Printf("Balance - %d\n", balance)
}
