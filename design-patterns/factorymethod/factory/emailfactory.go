package factory

import "designpatterns/factorymethod/product"

type EmailFactory struct{}

func (emailf *EmailFactory) Create() product.Notifier {
	return &product.Email{}
}
