package factory

import "designpatterns/factorymethod/product"

type SmsFactory struct{}

func (smsf *SmsFactory) Create() product.Notifier {
	return &product.Sms{}
}
