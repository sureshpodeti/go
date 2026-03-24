package main

import (
	"designpatterns/visitor/shape"
	"fmt"
)

func main() {
	square := shape.NewSquare(10)
	circle := shape.NewCircle(5)
	rectangle := shape.NewRectangle(10, 20)

	areaCalculator := shape.AreaCalculator{}
	square.Accept(&areaCalculator)
	circle.Accept(&areaCalculator)
	rectangle.Accept(&areaCalculator)

	fmt.Println()
	perimeterCalculator := &shape.PerimeterCalculator{}
	square.Accept(perimeterCalculator)
	circle.Accept(perimeterCalculator)
	rectangle.Accept(perimeterCalculator)
}
