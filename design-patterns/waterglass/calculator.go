package waterglass

type Calculator struct {
	Glass   *Glass
	Pricing *WaterPricing
}

func (c *Calculator) Calculate() float64 {
	waterInML := c.Glass.GetWater()
	priceForML := c.Pricing.GetPrice(1)
	return waterInML * priceForML
}
