package car

import (
	"fmt"
)

type Off struct {
	Car *Car
}

func (off *Off) EngineOff() {
	fmt.Println("Engine already off!")
}

func (off *Off) EngineOn() {
	off.Car.SetState(off.Car.onState)
	fmt.Println("Engine On!")
}

func (off *Off) Accelerate() {
	fmt.Println("Engine OFF, can not accelerate!")
}

func (off *Off) ApplyBrake() {
	fmt.Println("Engine OFF, applying brakes no effect!")

}

func (off *Off) RelaseBrake() {
	fmt.Println("Engine OFF, release brakes is no effect!")
}
