package shape

import "fmt"

type PerimeterCalculator struct{}

func (p *PerimeterCalculator) VisitForSquare(s *Square) {
	fmt.Println("Perimeter of Square:", 4*s.Side)
}

func (p *PerimeterCalculator) VisitForRectangle(r *Rectangle) {
	fmt.Println("Perimeter of Rectangle:", 2*(r.Width+r.Height))
}

func (p *PerimeterCalculator) VisitForCircle(c *Circle) {
	fmt.Println("Perimeter of Circle:", 2*3.14*c.Radius)
}