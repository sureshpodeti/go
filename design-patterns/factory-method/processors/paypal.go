package processors

import "fmt"

type payPal struct{ amount float64 }

func (p *payPal) Process() string {
	return fmt.Sprintf("charged %.2f via PayPal", p.amount)
}
