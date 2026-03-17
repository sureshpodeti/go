package decorator

type Margherita struct{}

func (m *Margherita) String() string {
	return "(Margherita)"
}

func (m *Margherita) Price() float64 {
	return 100.0
}
