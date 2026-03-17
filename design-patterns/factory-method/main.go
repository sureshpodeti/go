package main

import (
	"bufio"
	"designpatterns/factory-method/processors"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	options := []struct {
		label string
		key   string
	}{
		{"Credit Card", "credit"},
		{"PayPal", "paypal"},
	}

	fmt.Println("Select payment type:")
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

	p, err := processors.NewProcessor(options[choice-1].key, 99.99)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(p.Process())
}
