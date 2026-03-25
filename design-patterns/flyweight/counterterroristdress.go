package flyweight

type CounterTerroristDress struct {
	Color string
}

func NewCounterTerroristDress() *CounterTerroristDress {
	return &CounterTerroristDress{Color: "green"}
}
