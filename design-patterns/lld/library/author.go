package library

// Author represents a book author
// Separated from Book because authors can write multiple books
type Author struct {
	ID      string
	Name    string
	Country string
}

func NewAuthor(id, name, country string) *Author {
	return &Author{
		ID:      id,
		Name:    name,
		Country: country,
	}
}
