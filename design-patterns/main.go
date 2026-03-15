package main

import (
	"designpatterns/contexts"
	"designpatterns/strategies"
	"fmt"
	"log"
)

func main() {
	add := &strategies.Add{}
	subtract := &strategies.Subtract{}

	arc := contexts.NewArithmeticContext(add)

	result, err := arc.Execute(10, 20)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("arithmetic op result - %d\n", result)

	arc.SetStrategy(subtract)

	result, err = arc.Execute(10, 20)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("arithmetic op result - %d\n", result)
}
