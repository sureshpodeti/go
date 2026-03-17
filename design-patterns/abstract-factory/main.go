package main

import (
	"bufio"
	"designpatterns/abstract-factory/payments"
	"designpatterns/abstract-factory/payments/paypal"
	"designpatterns/abstract-factory/payments/stripe"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	options := []struct {
		label   string
		factory payments.Factory
	}{
		{"Stripe", &stripe.Factory{}},
		{"PayPal", &paypal.Factory{}},
	}

	fmt.Println("Select payment provider:")
	for i, o := range options {
		fmt.Printf("  %d. %s\n", i+1, o.label)
	}
	fmt.Print("Enter number: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(options) {
		log.Fatalf("invalid selection: %s", input)
	}

	factory := options[choice-1].factory

	processor := factory.CreateProcessor()
	refunder := factory.CreateRefunder()
	receiptGen := factory.CreateReceiptGenerator()

	fmt.Println(processor.Process(99.99))
	fmt.Println(refunder.Refund(25.00))
	fmt.Println(receiptGen.Generate(99.99))
}
