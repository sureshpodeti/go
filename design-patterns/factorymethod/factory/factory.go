package factory

import "designpatterns/factorymethod/product"

type Factory interface {
	Create() product.Notifier
}
