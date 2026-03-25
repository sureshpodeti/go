package main

import (
	"designpatterns/singleton"
	"fmt"
)

func main() {

	for i := 0; i < 30; i++ {
		go singleton.GetInstance()
	}

	fmt.Scanln()
}
