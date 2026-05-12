package car

import "fmt"

type Stop struct {
	Car *Car
}

func (sp *Stop) EngineOff() {
	sp.Car.SetState(sp.Car.offState)
	fmt.Println("Engine off!")
}

func (sp *Stop) EngineOn() {
	fmt.Println("Engine is already running!")
}

func (sp *Stop) Accelerate() {
	fmt.Println("First release brakes to accelerate vehicle")
}

func (sp *Stop) ApplyBrake() {
	fmt.Println("already brakes are applied")
}

func (sp *Stop) RelaseBrake() {
	sp.Car.SetState(sp.Car.movingState)
	fmt.Println("Car is moving!")
}
