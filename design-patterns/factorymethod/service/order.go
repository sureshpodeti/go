package service

import (
	"designpatterns/factorymethod/factory"
	"fmt"
)

type OrderService struct {
	F factory.Factory
}

func NewOrderService(f factory.Factory) *OrderService {
	return &OrderService{F: f}
}

func (o *OrderService) Order() {
	fmt.Println("Order created")

	//Notify about order
	channel := o.F.Create()

	channel.Notify()
}
