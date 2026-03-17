package main

import (
	"bufio"
	"designpatterns/dependency-injection/client"
	"fmt"
	"log"
	"os"
	"strings"
)

func fetchData(c client.Client, url string) {
	fmt.Println(c.Get(url))
}

func main() {
	options := []struct {
		label   string
		useMock bool
	}{
		{"Production (HTTP)", false},
		{"Mock (Testing)", true},
	}

	fmt.Println("Select client mode:")
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

	c := client.NewClient(options[choice-1].useMock)

	fetchData(c, "https://api.example.com/users")
	fetchData(c, "https://api.example.com/orders")
}
