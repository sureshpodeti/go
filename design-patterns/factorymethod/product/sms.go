package product

import "fmt"

type Sms struct{}

func (sms *Sms) Notify() {
	fmt.Println("Sending Sms!")
}
