package composite

import "fmt"

type Folder struct {
	Name     string
	Children []Component
}

func NewFolder(name string) *Folder {
	return &Folder{
		Name:     name,
		Children: make([]Component, 0),
	}
}

func (fld *Folder) Add(c Component) {
	fld.Children = append(fld.Children, c)
}

func (fld *Folder) Search(keyword string) {
	fmt.Printf("Search '%s' in folder '%s'\n", keyword, fld.Name)
	for _, child := range fld.Children {
		child.Search(keyword)
	}
}
