package main

import "designpatterns/bridge"

func main() {
	hpPrinter := &bridge.Hp{}
	epsonPrinter := &bridge.Epson{}

	mac := bridge.NewMac(hpPrinter)
	mac.Print()

	mac.SetPrinter(epsonPrinter)
	mac.Print()

	windows := bridge.NewWindows(epsonPrinter)
	windows.Print()

	windows.SetPrinter(hpPrinter)
	windows.Print()
}
