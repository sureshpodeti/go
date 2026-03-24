package composite

import "fmt"

type File struct {
	Name string
}

func NewFile(name string) *File {
	return &File{Name: name}
}

func (fl *File) Search(keyword string) {
	fmt.Printf("Searching for keyword %s in file %s\n", keyword, fl.Name)
}
