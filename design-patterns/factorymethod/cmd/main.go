package main

import (
	"designpatterns/factorymethod/factory"
	"designpatterns/factorymethod/service"
)

func main() {

	notificationType := "sms"

	var f factory.Factory

	switch notificationType {
	case "sms":
		f = &factory.SmsFactory{}
	case "email":
		f = &factory.EmailFactory{}
	}

	authService := service.NewAuthService(f)
	authService.Login()

	orderService := service.NewOrderService(f)
	orderService.Order()

}
