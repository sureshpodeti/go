package waterglass1

type Calculator struct {
	Container Container
}

func (c *Calculator) Calculate() float64 {
	liquid := c.Container.GetLiquid()
	return c.Container.GetLiquidVolume() * liquid.GetUnitPrice()
}
