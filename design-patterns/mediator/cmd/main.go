package main

import (
	"designpatterns/mediator"
)

func main() {
	stationManager := mediator.NewStationManager()

	passagnerTrain := mediator.NewPassangerTrain(stationManager)
	frieghtTrain := mediator.NewFrieghtTrain(stationManager)

	passagnerTrain.Arrive()
	frieghtTrain.Arrive()
	passagnerTrain.Departure()
}
