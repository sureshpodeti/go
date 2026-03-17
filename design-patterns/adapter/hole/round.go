package hole

type Peg interface {
	Radius() float64
}

type RoundHole struct {
	Radius float64
}

func (rh *RoundHole) Fits(peg Peg) bool {
	return peg.Radius() <= rh.Radius
}
