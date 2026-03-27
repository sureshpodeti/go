package service

import (
	"designpatterns/builder1"
	"designpatterns/builder1/builder"
	"designpatterns/builder1/product"
)

type SuburbService struct {
	Director *builder1.Director
	Builder  *builder.HouseBuilder
}

func (s *SuburbService) CreateHouse() *product.House {
	s.Director.SetBuilder(s.Builder)
	s.Director.BuildHouse()
	return s.Builder.GetHouse()
}
