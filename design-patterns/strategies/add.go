package strategies

type Add struct{}

func (a *Add) Do(a1, a2 int) int {
	return a1 + a2
}
