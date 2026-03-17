package decorator

type Tomato struct {
	pizza Pizza
}

func NewTomato(p Pizza) *Tomato {
	return &Tomato{pizza: p}
}

func (t *Tomato) String() string {
	return t.pizza.String() + " + (Tomato)"
}

func (t *Tomato) Price() float64 {
	return t.pizza.Price() + 30.0
}
