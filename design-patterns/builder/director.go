package builder

type Director struct {
	builder IBuilder
}

func NewDirector(b IBuilder) *Director {
	return &Director{builder: b}
}

func (d *Director) SetBuilder(b IBuilder) {
	d.builder = b
}

func (d *Director) BuildHouse() House {
	d.builder.setWindowType()
	d.builder.setDoorType()
	d.builder.setFloor()
	return d.builder.getHouse()
}
