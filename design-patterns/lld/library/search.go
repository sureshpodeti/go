package library

import (
	"strings"
	"time"
)

// SearchStrategy defines interface for different search strategies
// Using Strategy Pattern to handle different search criteria
type SearchStrategy interface {
	Search(books []*Book) []*Book
}

// TitleSearchStrategy searches books by title
type TitleSearchStrategy struct {
	Title string
}

func (s *TitleSearchStrategy) Search(books []*Book) []*Book {
	result := make([]*Book, 0)
	searchTerm := strings.ToLower(s.Title)
	for _, book := range books {
		if strings.Contains(strings.ToLower(book.Title), searchTerm) {
			result = append(result, book)
		}
	}
	return result
}

// AuthorSearchStrategy searches books by author name
type AuthorSearchStrategy struct {
	AuthorName string
}

func (s *AuthorSearchStrategy) Search(books []*Book) []*Book {
	result := make([]*Book, 0)
	searchTerm := strings.ToLower(s.AuthorName)
	for _, book := range books {
		for _, author := range book.Authors {
			if strings.Contains(strings.ToLower(author.Name), searchTerm) {
				result = append(result, book)
				break
			}
		}
	}
	return result
}

// SubjectSearchStrategy searches books by subject
type SubjectSearchStrategy struct {
	Subject string
}

func (s *SubjectSearchStrategy) Search(books []*Book) []*Book {
	result := make([]*Book, 0)
	searchTerm := strings.ToLower(s.Subject)
	for _, book := range books {
		if strings.Contains(strings.ToLower(book.Subject), searchTerm) {
			result = append(result, book)
		}
	}
	return result
}

// PublicationDateSearchStrategy searches books by publication date range
type PublicationDateSearchStrategy struct {
	StartDate time.Time
	EndDate   time.Time
}

func (s *PublicationDateSearchStrategy) Search(books []*Book) []*Book {
	result := make([]*Book, 0)
	for _, book := range books {
		if (book.PublicationDate.Equal(s.StartDate) || book.PublicationDate.After(s.StartDate)) &&
			(book.PublicationDate.Equal(s.EndDate) || book.PublicationDate.Before(s.EndDate)) {
			result = append(result, book)
		}
	}
	return result
}

// Catalog manages book searching
// Responsibility: Provide search functionality
// Why not in Book? Because searching requires knowledge of ALL books
type Catalog struct {
	books []*Book
}

func NewCatalog() *Catalog {
	return &Catalog{
		books: make([]*Book, 0),
	}
}

func (c *Catalog) AddBook(book *Book) {
	c.books = append(c.books, book)
}

func (c *Catalog) Search(strategy SearchStrategy) []*Book {
	return strategy.Search(c.books)
}
