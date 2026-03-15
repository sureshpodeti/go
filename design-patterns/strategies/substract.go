package strategies

type Subtract struct{}

func (s *Subtract) Do(a1, a2 int) int {
	return a1 - a2
}
