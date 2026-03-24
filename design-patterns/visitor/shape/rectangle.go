package shape

type Rectangle struct {
	Width, Height float64
}

func NewRectangle(width, height float64) *Rectangle {
	return &Rectangle{width, height}
}

func (r *Rectangle) Accept(v Visitor) {
	v.VisitForRectangle(r)
}
