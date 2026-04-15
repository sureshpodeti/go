package waterglass

type Water struct {
	Ph float64
}

func (w *Water) GetPh() float64 {
	return w.Ph
}
