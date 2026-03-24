package mediator

type StationManager struct {
	IsPlatformFree bool
	Trains         []Train
}

func NewStationManager() *StationManager {
	return &StationManager{
		IsPlatformFree: true,
	}
}

func (sm *StationManager) CanArrive(t Train) bool {
	if sm.IsPlatformFree {
		sm.IsPlatformFree = false
		return true
	}
	sm.Trains = append(sm.Trains, t)
	return false
}

func (sm *StationManager) NotifyDeparture() {
	if !sm.IsPlatformFree {
		sm.IsPlatformFree = true
	}

	if len(sm.Trains) > 0 {
		firstTrainInQueue := sm.Trains[0]
		sm.Trains = sm.Trains[1:]
		firstTrainInQueue.PermitArrival()
	}

}
