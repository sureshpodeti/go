package department

import "designpatterns/chainofresponsibility/model"

type Cashier struct {
	Next Department
}

func (c *Cashier) Executer(p *model.Patient) {
	if p.PaymentDone {
		println("Payment already done")
	} else {
		println("Cashier getting money from patient")
		p.PaymentDone = true
	}
}
