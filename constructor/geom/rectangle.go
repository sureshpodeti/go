package geom

type Rectangle struct {
	length float64
	width  float64
}

func NewRectangle(opts ...Option) *Rectangle {
	r := &Rectangle{
		length: 1.0,
		width:  1.0,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type Option func(*Rectangle)

func WithLength(l float64) Option {
	return func(r *Rectangle) {
		r.length = l
	}
}

func WithWidth(w float64) Option {
	return func(r *Rectangle) {
		r.width = w
	}
}

func (r *Rectangle) Area() float64 {
	return r.length * r.width
}

func (r *Rectangle) Perimter() float64 {
	return 2 * (r.length + r.width)
}
