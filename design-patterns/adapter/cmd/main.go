package main

import (
	"designpatterns/adapter/adapter"
	"designpatterns/adapter/hole"
	"designpatterns/adapter/peg"
	"fmt"
)

func main() {
	rhole := &hole.RoundHole{Radius: 5}

	roundPeg := &peg.RoundPeg{R: 3}
	fmt.Printf("RoundPeg fits the hole - %t\n", rhole.Fits(roundPeg))

	squarePeg := &peg.SquarePeg{W: 6}
	sqAdapter := &adapter.SquarePegAdapter{Sqp: squarePeg}
	fmt.Printf("SquarePeg fits the hole - %t\n", rhole.Fits(sqAdapter))
}
