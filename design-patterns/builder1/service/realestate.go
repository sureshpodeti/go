package service

import (
	"designpatterns/builder1"
	"designpatterns/builder1/builder"
	"fmt"
)

type RealEstateService struct {
	Director     *builder1.Director
	IglooBuilder *builder.IglooBuilder
	HouseBuilder *builder.HouseBuilder
}

func (r *RealEstateService) BuildHouse() {
	r.Director.SetBuilder(r.IglooBuilder)
	r.Director.BuildIgloo()
	igloo := r.IglooBuilder.GetHouse()
	fmt.Println("Igloo - ", igloo)

	r.Director.SetBuilder(r.HouseBuilder)
	r.Director.BuildHouse()
	house := r.HouseBuilder.GetHouse()
	fmt.Println("House - ", house)
}
