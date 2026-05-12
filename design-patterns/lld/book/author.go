package main

// Author represents a book author (immutable)
type Author struct {
	id   string
	name string
}

// NewAuthor creates a new author
func NewAuthor(id, name string) *Author {
	return &Author{
		id:   id,
		name: name,
	}
}

// GetID returns the author's unique identifier
func (a *Author) GetID() string {
	return a.id
}

// GetName returns the author's name
func (a *Author) GetName() string {
	return a.name
}
