package main

import (
	"designpatterns/builder"
	"designpatterns/builder/constant"
	"fmt"
)

func main() {
	igloo := builder.GetBuilder(constant.IGLOO)
	normal := builder.GetBuilder(constant.NORMAL)

	director := builder.NewDirector(normal)

	normalHouse := director.BuildHouse()

	fmt.Printf("Normal House Door Type: %s\n", normalHouse.DoorType)
	fmt.Printf("Normal House Window Type: %s\n", normalHouse.WindowType)
	fmt.Printf("Normal House Num Floor: %d\n", normalHouse.Floor)

	director.SetBuilder(igloo)

	iglooHouse := director.BuildHouse()

	fmt.Printf("\nIgloo House Door Type: %s\n", iglooHouse.DoorType)
	fmt.Printf("Igloo House Window Type: %s\n", iglooHouse.WindowType)
	fmt.Printf("Igloo House Num Floor: %d\n", iglooHouse.Floor)

}
