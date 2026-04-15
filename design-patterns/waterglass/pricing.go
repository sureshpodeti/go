package waterglass

type WaterPricing struct {
	PriceForML float64
}

func (wp *WaterPricing) GetPrice(ml int) float64 {
	return float64(ml) * wp.PriceForML
}
