package peg

type SquarePeg struct {
	W float64
}

func (s *SquarePeg) Width() float64 {
	return s.W
}
