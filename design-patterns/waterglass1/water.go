package waterglass1

type Water struct {
	PriceForML float64
}

func (w *Water) GetUnitPrice() float64 {
	return w.PriceForML
}
