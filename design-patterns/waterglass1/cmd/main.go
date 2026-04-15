package main

import (
	"designpatterns/waterglass1"
	"fmt"
)

func main() {
	glass := waterglass1.NewGlass(20000)

	water := &waterglass1.Water{
		PriceForML: 0.0005,
	}

	glass.PourLiquid(water, 40000)

	calculator := &waterglass1.Calculator{Container: glass}

	fmt.Println("Calculator - ", calculator.Calculate())

}
