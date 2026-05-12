package main

import (
	"designpatterns/lld/car/car"
)

func main() {
	car := car.NewCar(car.BLACK, car.SUV)

	car.EngineOn()

	car.Accelerate()

	car.ApplyBrake()

	car.RelaseBrake()

	car.ApplyBrake()

	car.EngineOff()

}
