package processors

import "fmt"

type Processor interface {
	Process() string
}

func NewProcessor(method string, amount float64) (Processor, error) {
	switch method {
	case "credit":
		return &creditCard{amount: amount}, nil
	case "paypal":
		return &payPal{amount: amount}, nil
	default:
		return nil, fmt.Errorf("unknown payment method: %s", method)
	}
}
