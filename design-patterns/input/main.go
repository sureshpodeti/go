package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	//Take single integer input
	var n int
	fmt.Print("Enter integer - ")
	fmt.Scanln(&n)
	fmt.Printf("You have entered %d\n", n)

	// Take single string input
	var str string
	fmt.Print("Enter string: ")
	fmt.Scanln(&str)
	fmt.Printf("You entered %s\n", str)

	//Take two integer inputs
	var a, b int
	fmt.Print("Enter a, b: ")
	fmt.Scan(&a, &b)
	fmt.Printf("sum = %d\n", a+b)

	// Read array of intergers
	var ar [5]int
	fmt.Println("Enter 5 integers")
	for i := 0; i < 5; i++ {
		fmt.Scan(&ar[i])
	}
	for _, e := range ar {
		fmt.Println(e)
	}

	// Read entire line
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter a line of words: ")
	scanner.Scan()
	line := scanner.Text()
	fmt.Println("line ", line)
	fmt.Println("Split by white space into slice - ", strings.Fields(line))

	// Split line by comma(,)
	scnner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter a line of words: ")
	scnner.Scan()
	txt := scnner.Text()
	splitresult := strings.Split(txt, ", ")
	fmt.Println(splitresult)

	
}
