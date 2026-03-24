package bridge

import "fmt"

type Epson struct{}

func (eps *Epson) PrintFile() {
	fmt.Println("Printing from epson!")
}
