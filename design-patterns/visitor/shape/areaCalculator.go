package shape

import "fmt"

type AreaCalculator struct{}

func (a *AreaCalculator) VisitForSquare(s *Square) {
	fmt.Printf("Square are - %f\n", s.Side*s.Side)
}

func (a *AreaCalculator) VisitForRectangle(r *Rectangle) {
	fmt.Printf("Rectangle area - %f\n", r.Width*r.Height)
}

func (a *AreaCalculator) VisitForCircle(c *Circle) {
	fmt.Printf("Circle area - %f\n", 3.14*c.Radius*c.Radius)
}
