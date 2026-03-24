package bridge

import "fmt"

type Mac struct {
	Printer Printer
}

func NewMac(printer Printer) *Mac {
	return &Mac{Printer: printer}
}

func (mac *Mac) SetPrinter(printer Printer) {
	mac.Printer = printer
}

func (mac *Mac) Print() {
	fmt.Println("Print request for mac!")
	mac.Printer.PrintFile()
}
