package bridge

import "fmt"

type Hp struct{}

func (hp *Hp) PrintFile() {
	fmt.Println("Printing by Hp printer!")
}
