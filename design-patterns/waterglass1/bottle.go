package waterglass1

type Bottle struct {
	capacityInML     float64
	liquid           *Liquid
	liquidVolumeInML float64
}

func NewBottle(vol float64) *Bottle {
	return &Bottle{capacityInML: vol}
}

func (b *Bottle) PourLiquid(lq Liquid, ml float64) {
	if b.liquid != nil {
		b.Empty()
	}
	b.liquid = &lq
	b.liquidVolumeInML = ml
}

func (b *Bottle) GetLiquid() Liquid {
	return *b.liquid
}

func (b *Bottle) Empty() {
	b.liquid = nil
	b.liquidVolumeInML = 0
}
