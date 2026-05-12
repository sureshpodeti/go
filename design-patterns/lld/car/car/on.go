package car

import "fmt"

type On struct {
	Car *Car
}

func (on *On) EngineOff() {
	on.Car.SetState(on.Car.offState)
	fmt.Println("Engine Off!")

}

func (on *On) EngineOn() {
	fmt.Println("Engine already on!")
}

func (on *On) Accelerate() {
	on.Car.SetState(on.Car.movingState)
	fmt.Println("Car Accelerating!")
}

func (on *On) ApplyBrake() {
	
	fmt.Println("Engine on!, and stopped. Your applied brakes!")
}

func (on *On) RelaseBrake() {
	fmt.Println("Engine on!, and stopped. Your release brakes has no effect!")
}
