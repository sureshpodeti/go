package peg

type RoundPeg struct {
	R float64
}

func (r *RoundPeg) Radius() float64 {
	return r.R
}
