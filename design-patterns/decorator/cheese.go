package decorator

type Cheese struct {
	pizza Pizza
}

func NewCheese(p Pizza) *Cheese {
	return &Cheese{pizza: p}
}

func (c *Cheese) String() string {
	return c.pizza.String() + " + (Cheese)"
}

func (c *Cheese) Price() float64 {
	return c.pizza.Price() + 60.0
}
