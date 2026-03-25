package product

import "fmt"

type Email struct{}

func (email *Email) Notify() {
	fmt.Println("Sending email!")
}
