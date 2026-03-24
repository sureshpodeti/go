package mediator

type Mediator interface {
	CanArrive(t Train) bool
	NotifyDeparture()
}
