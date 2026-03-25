package flyweight

type Player struct {
	Type      string
	Lat, Long int
}

func NewPlayer(playerType string) *Player {
	return &Player{
		Type: playerType,
	}
}
