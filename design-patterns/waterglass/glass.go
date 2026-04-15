package waterglass

type Glass struct {
	SizeInML  float64
	WaterInML float64
}

func NewGlass(size float64) *Glass {
	return &Glass{SizeInML: size, WaterInML: 0}
}

func (g *Glass) AddWater(ml float64) {
	if ml > g.SizeInML {
		g.WaterInML = g.SizeInML
	} else {
		g.WaterInML = ml
	}
}

func (g *Glass) GetWater() float64 {
	return g.WaterInML
}
