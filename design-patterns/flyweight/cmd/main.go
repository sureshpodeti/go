package main

import (
	"designpatterns/flyweight"
)

func main() {
	game := flyweight.NewGame()

	player1 := game.AddCounterTerrorist()
}
