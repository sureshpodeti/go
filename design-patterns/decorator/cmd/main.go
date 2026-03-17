package main

import (
	"designpatterns/decorator"
	"fmt"
)

func main() {
	var pizza decorator.Pizza = &decorator.Margherita{}

	pizza = decorator.NewMushroom(pizza)
	pizza = decorator.NewTomato(pizza)
	pizza = decorator.NewCheese(pizza)

	fmt.Printf("Item - %s, price - %.2f\n", pizza, pizza.Price())
}
