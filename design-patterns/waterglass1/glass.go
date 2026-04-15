package waterglass1

type Glass struct {
	capacityInML     float64
	liquid           *Liquid
	liquidVolumeInML float64
}

func NewGlass(vol float64) *Glass {
	return &Glass{capacityInML: vol}
}

func (g *Glass) PourLiquid(lq Liquid, ml float64) {
	if g.liquid != nil {
		g.Empty()
	}
	g.liquid = &lq
	g.liquidVolumeInML = ml
}

func (g *Glass) GetLiquid() Liquid {
	return *g.liquid
}

func (g *Glass) GetLiquidVolume() float64 {
	return g.liquidVolumeInML
}

func (g *Glass) Empty() {
	g.liquid = nil
	g.liquidVolumeInML = 0
}
