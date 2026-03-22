package device

import "fmt"

type Tv struct {
	isRunning bool
}

func NewTv() *Tv {
	return &Tv{}
}

func (tv *Tv) On() {
	tv.isRunning = true
	fmt.Println("Tv is On!")
}

func (tv *Tv) Off() {
	tv.isRunning = false
	fmt.Println("Tv is Off!")
}
