package main

import "fmt"

func main() {
	ch := make(chan bool)

	<-ch

	fmt.Println("Hello World")
}
