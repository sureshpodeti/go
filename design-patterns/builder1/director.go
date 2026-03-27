package builder1

import "designpatterns/builder1/builder"

type Director struct {
	Builder builder.BuildingBuilder
}

func NewDirector(b builder.BuildingBuilder) *Director {
	return &Director{Builder: b}
}

func (d *Director) SetBuilder(b builder.BuildingBuilder) {
	d.Builder = b
}

func (d *Director) BuildHouse() {
	d.Builder.Reset()
	d.Builder.BuildWalls()
	d.Builder.BuildDoor()
}

func (d *Director) BuildIgloo() {
	d.Builder.Reset()
	d.Builder.BuildWalls()
	d.Builder.BuildDoor()
	d.Builder.BuildRoof()
}
