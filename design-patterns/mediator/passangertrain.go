package mediator

import "fmt"

type PassangerTrain struct {
	Mediator Mediator
}

func NewPassangerTrain(mediator Mediator) *PassangerTrain {
	return &PassangerTrain{Mediator: mediator}
}

func (p *PassangerTrain) Arrive() {
	if !p.Mediator.CanArrive(p) {
		fmt.Println("PassengerTrain: Arrival blocked, waiting")
		return
	}
	fmt.Println("PassengerTrain: Arrived")

}

func (p *PassangerTrain) Departure() {
	fmt.Println("PassengerTrain: Leaving")
	p.Mediator.NotifyDeparture()
}

func (p *PassangerTrain) PermitArrival() {
	fmt.Println("PassengerTrain: Arrival permitted, arriving")
	p.Arrive()
}
