package main

import (
	"designpatterns/waterglass"
	"fmt"
)

func main() {
	glass := waterglass.NewGlass(10000)

	glass.AddWater(1000)

	pricing := &waterglass.WaterPricing{PriceForML: 0.001}

	calculator := &waterglass.Calculator{Pricing: pricing, Glass: glass}

	fmt.Println("Calculator - ", calculator.Calculate())

}
