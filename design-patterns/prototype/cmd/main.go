package main

import (
	"designpatterns/prototype"
	"fmt"
)

func Duplicate(p prototype.Prototype) prototype.Prototype {
	return p.Clone()
}
func main() {
	original := &prototype.TemplatedDocument{
		Document:     prototype.Document{Title: "Report", Content: "Q4 results", Author: "Alice"},
		TemplateName: "quarterly",
		Version:      3,
	}

	copy := Duplicate(original)

	fmt.Printf("Original: %+v\n", original)
	fmt.Printf("Copy:     %+v\n", copy)

	// They're different objects
	fmt.Println("Same pointer?", original == copy) // false

}
