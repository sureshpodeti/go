package main

import "designpatterns/composite"

func main() {

	file1 := composite.NewFile("file 1")
	file2 := composite.NewFile("file 2")
	file3 := composite.NewFile("file 3")

	folder1 := composite.NewFolder("folder 1")
	folder1.Add(file1)

	folder2 := composite.NewFolder("folder 2")
	folder2.Add(file2)
	folder2.Add(file3)

	folder2.Add(folder1)

	folder2.Search("rose")

}
