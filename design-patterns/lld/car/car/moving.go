package car

import "fmt"

type Moving struct {
	Car *Car
}

func (mv *Moving) EngineOff() {
	fmt.Println("You can't stop the vehicle while vehicle is running")
}

func (mv *Moving) EngineOn() {
	fmt.Println("Engine is on and car is running. You can't switch on the engine")
}

func (mv *Moving) Accelerate() {
	fmt.Println("Speed increased")
}

func (mv *Moving) ApplyBrake() {
	mv.Car.SetState(mv.Car.stopState)
	fmt.Println("Car is stopped")

}

func (mv *Moving) RelaseBrake() {
	fmt.Println("Moving vehicle has no effect on release brake!")
}
