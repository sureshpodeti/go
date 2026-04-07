package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	var result []string
	//Open file - file1
	file1, err := os.Open("file1.txt")
	if err != nil {
		fmt.Println("Failed to open file")
	}
	defer file1.Close()

	//Open file - file2
	file2, err := os.Open("file2.txt")
	if err != nil {
		fmt.Println("Failed to open file")
	}
	defer file2.Close()

	buf1 := bufio.NewScanner(file1)
	buf2 := bufio.NewScanner(file2)

	has1, has2 := buf1.Scan(), buf2.Scan()

	for has1 && has2 {
		if buf1.Text() <= buf2.Text() {
			result = append(result, buf1.Text())
			has1 = buf1.Scan()
		} else {
			result = append(result, buf2.Text())
			has2 = buf2.Scan()
		}
	}

	//Write to file
	resultFile, err := os.Create("result.txt")
	if err != nil {
		fmt.Println("Unable to create a file")
	}
	defer resultFile.Close()

	writer := bufio.NewWriter(resultFile)

	for _, line := range result {
		writer.WriteString(line)
		writer.WriteByte('\n')
	}

	writer.Flush()
}
