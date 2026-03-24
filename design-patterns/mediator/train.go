package mediator

type Train interface {
	Arrive()
	Departure()
	PermitArrival()
}
