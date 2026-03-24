package shape

type Square struct {
	Side float64
}

func NewSquare(s float64) *Square {
	return &Square{Side: s}
}
func (s *Square) Accept(v Visitor) {
	v.VisitForSquare(s)
}
