package decorator

type Mushroom struct {
	pizza Pizza
}

func NewMushroom(p Pizza) *Mushroom {
	return &Mushroom{pizza: p}
}

func (m *Mushroom) String() string {
	return m.pizza.String() + " + (Mushroom)"
}

func (m *Mushroom) Price() float64 {
	return m.pizza.Price() + 50.0
}
