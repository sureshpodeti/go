package service

import (
	"designpatterns/factorymethod/factory"
	"fmt"
)

type AuthService struct {
	F factory.Factory
}

func NewAuthService(f factory.Factory) *AuthService {
	return &AuthService{F: f}
}

func (auth *AuthService) Login() {
	fmt.Println("Successful login!")
	channel := auth.F.Create()
	channel.Notify()
}
