package library

import "time"

// Book represents the metadata of a book (not a physical copy)
// One Book can have multiple BookItems (physical copies)
// Book is responsible for: knowing its metadata
// Book is NOT responsible for: searching, availability, location
type Book struct {
	ISBN            string
	Title           string
	Authors         []*Author
	Subject         string
	PublicationDate time.Time
	NumPages        int
}

func NewBook(isbn, title, subject string, pubDate time.Time, numPages int) *Book {
	return &Book{
		ISBN:            isbn,
		Title:           title,
		Subject:         subject,
		PublicationDate: pubDate,
		NumPages:        numPages,
		Authors:         make([]*Author, 0),
	}
}

func (b *Book) AddAuthor(author *Author) {
	b.Authors = append(b.Authors, author)
}

// GetAuthorNames returns comma-separated author names
func (b *Book) GetAuthorNames() string {
	if len(b.Authors) == 0 {
		return ""
	}
	names := ""
	for i, author := range b.Authors {
		if i > 0 {
			names += ", "
		}
		names += author.Name
	}
	return names
}
