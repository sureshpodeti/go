package bridge

import "fmt"

type Windows struct {
	Printer Printer
}

func NewWindows(p Printer) *Windows {
	return &Windows{Printer: p}
}

func (w *Windows) SetPrinter(printer Printer) {
	w.Printer = printer
}

func (w *Windows) Print() {
	fmt.Println("Print request for windows!")
	w.Printer.PrintFile()
}
