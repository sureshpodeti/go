package main

import (
	"designpatterns/builder1"
	"designpatterns/builder1/builder"
	"designpatterns/builder1/service"
	"fmt"
)

func main() {
	iglooBuilder := &builder.IglooBuilder{}
	houseBuilder := &builder.HouseBuilder{}

	director := builder1.NewDirector(iglooBuilder)

	// Services get what they need
	arcticService := &service.ArcticService{
		Director: director,
		Builder:  iglooBuilder,
	}

	suburbService := &service.SuburbService{
		Director: director,
		Builder:  houseBuilder,
	}

	realestateService := &service.RealEstateService{
		IglooBuilder: iglooBuilder,
		HouseBuilder: houseBuilder,
		Director:     director,
	}

	igloo := arcticService.CreateHouse()
	fmt.Println("igloo - ", igloo)
	house := suburbService.CreateHouse()
	fmt.Println("house - ", house)

	realestateService.BuildHouse()
}
