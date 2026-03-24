package shape

type Circle struct {
	Radius float64
}

func NewCircle(r float64) *Circle {
	return &Circle{Radius: r}
}

func (c *Circle) Accept(v Visitor) {
	v.VisitForCircle(c)
}
