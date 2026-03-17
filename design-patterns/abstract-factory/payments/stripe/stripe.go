package stripe

import (
	"designpatterns/abstract-factory/payments"
	"fmt"
)

type Factory struct{}

func (f *Factory) CreateProcessor() payments.Processor             { return &processor{} }
func (f *Factory) CreateRefunder() payments.Refunder               { return &refunder{} }
func (f *Factory) CreateReceiptGenerator() payments.ReceiptGenerator { return &receipt{} }

type processor struct{}

func (p *processor) Process(amount float64) string {
	return fmt.Sprintf("stripe: charged %.2f", amount)
}

type refunder struct{}

func (r *refunder) Refund(amount float64) string {
	return fmt.Sprintf("stripe: refunded %.2f", amount)
}

type receipt struct{}

func (r *receipt) Generate(amount float64) string {
	return fmt.Sprintf("stripe: receipt for %.2f", amount)
}
