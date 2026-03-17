package adapter

import (
	pegs "designpatterns/adapter/peg"
	"math"
)

type SquarePegAdapter struct {
	Sqp *pegs.SquarePeg
}

func (spa *SquarePegAdapter) Radius() float64 {
	return spa.Sqp.Width() * math.Sqrt(2) / 2
}
