package service

import (
	"designpatterns/builder1"
	"designpatterns/builder1/builder"
	"designpatterns/builder1/product"
)

type ArcticService struct {
	Director *builder1.Director
	Builder  *builder.IglooBuilder
}

func (a *ArcticService) CreateHouse() *product.Igloo {
	a.Director.SetBuilder(a.Builder)
	a.Director.BuildHouse()
	return a.Builder.GetHouse()
}
