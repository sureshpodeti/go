package flyweight

type Game struct {
	Terrorists        []*Player
	CounterTerrorists []*Player
}

func NewGame() *Game {
	return &Game{
		Terrorists:        make([]*Player, 0),
		CounterTerrorists: make([]*Player, 0),
	}
}

func (g *Game) AddTerrorist(p *Player) {
	g.Terrorists = append(g.Terrorists, p)
}

func (g *Game) AddCounterTerrorist(p *Player) {
	g.CounterTerrorists = append(g.CounterTerrorists, p)
}
