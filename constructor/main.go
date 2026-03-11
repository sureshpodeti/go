package main

import (
	"fmt"
	"practise/constructor/geom"
)

func main() {
	r := geom.NewRectangle(
		geom.WithLength(20.50),
		geom.WithWidth(100.50),
	)

	fmt.Printf("Area - %.2f, Perimter - %.2f\n", r.Area(), r.Perimter())
}
