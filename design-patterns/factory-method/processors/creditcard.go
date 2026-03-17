package processors

import "fmt"

type creditCard struct{ amount float64 }

func (c *creditCard) Process() string {
	return fmt.Sprintf("charged %.2f to credit card", c.amount)
}
