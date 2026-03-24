package mediator

import "fmt"

type FrieghtTrain struct {
	Mediator Mediator
}

func NewFrieghtTrain(mediator Mediator) *FrieghtTrain {
	return &FrieghtTrain{Mediator: mediator}
}

func (ft *FrieghtTrain) Arrive() {
	if !ft.Mediator.CanArrive(ft) {
		fmt.Println("FreightTrain: Arrival blocked, waiting")
		return
	}
	fmt.Println("FreightTrain: Arrived")
}

func (ft *FrieghtTrain) Departure() {
	fmt.Println("FreightTrain: Leaving")
	ft.Mediator.NotifyDeparture()
}

func (ft *FrieghtTrain) PermitArrival() {
	fmt.Println("FreightTrain: Arrival permitted")
	ft.Arrive()
}
