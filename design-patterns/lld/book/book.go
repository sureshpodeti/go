package main

import (
	"errors"
	"fmt"
)

// Book represents a book with pages and authors
type Book struct {
	isbn            string
	title           string
	publicationYear int
	authors         []*Author
	pages           []*Page
	currentPageIdx  int
	isOpen          bool
}

// NewBook creates a new book
func NewBook(isbn, title string, publicationYear int) *Book {
	return &Book{
		isbn:            isbn,
		title:           title,
		publicationYear: publicationYear,
		authors:         make([]*Author, 0),
		pages:           make([]*Page, 0),
		currentPageIdx:  0,
		isOpen:          false,
	}
}

// GetISBN returns the book's ISBN
func (b *Book) GetISBN() string {
	return b.isbn
}

// GetTitle returns the book's title
func (b *Book) GetTitle() string {
	return b.title
}

// GetPublicationYear returns the publication year
func (b *Book) GetPublicationYear() int {
	return b.publicationYear
}

// GetAuthors returns all authors
func (b *Book) GetAuthors() []*Author {
	return b.authors
}

// AddAuthor adds an author to the book
func (b *Book) AddAuthor(author *Author) error {
	if author == nil {
		return errors.New("author cannot be nil")
	}
	b.authors = append(b.authors, author)
	return nil
}

// RemoveAuthor removes an author by ID
func (b *Book) RemoveAuthor(authorID string) error {
	for i, author := range b.authors {
		if author.GetID() == authorID {
			b.authors = append(b.authors[:i], b.authors[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("author with ID %s not found", authorID)
}

// AddPage adds a page to the book
func (b *Book) AddPage(page *Page) error {
	if page == nil {
		return errors.New("page cannot be nil")
	}
	b.pages = append(b.pages, page)
	return nil
}

// GetNumberOfPages returns the total number of pages
func (b *Book) GetNumberOfPages() int {
	return len(b.pages)
}

// Open opens the book for reading
func (b *Book) Open() error {
	if b.isOpen {
		return errors.New("book is already open")
	}
	if len(b.pages) == 0 {
		return errors.New("book has no pages")
	}
	b.isOpen = true
	b.currentPageIdx = 0
	return nil
}

// Close closes the book
func (b *Book) Close() error {
	if !b.isOpen {
		return errors.New("book is already closed")
	}
	b.isOpen = false
	b.currentPageIdx = 0
	return nil
}

// IsBookOpen returns whether the book is open
func (b *Book) IsBookOpen() bool {
	return b.isOpen
}

// NextPage moves to the next page
func (b *Book) NextPage() error {
	if !b.isOpen {
		return errors.New("book is closed")
	}
	if b.currentPageIdx >= len(b.pages)-1 {
		return errors.New("already at last page")
	}
	b.currentPageIdx++
	return nil
}

// PrevPage moves to the previous page
func (b *Book) PrevPage() error {
	if !b.isOpen {
		return errors.New("book is closed")
	}
	if b.currentPageIdx <= 0 {
		return errors.New("already at first page")
	}
	b.currentPageIdx--
	return nil
}

// GoToPage jumps to a specific page number
func (b *Book) GoToPage(pageNum int) error {
	if !b.isOpen {
		return errors.New("book is closed")
	}
	idx := pageNum - 1 // Convert to 0-based index
	if idx < 0 || idx >= len(b.pages) {
		return fmt.Errorf("invalid page number: %d", pageNum)
	}
	b.currentPageIdx = idx
	return nil
}

// GetCurrentPageNumber returns the current page number
func (b *Book) GetCurrentPageNumber() int {
	if !b.isOpen {
		return 0
	}
	return b.currentPageIdx + 1 // Convert to 1-based
}

// GetCurrentPage returns the current page object
func (b *Book) GetCurrentPage() (*Page, error) {
	if !b.isOpen {
		return nil, errors.New("book is closed")
	}
	return b.pages[b.currentPageIdx], nil
}

// ReadCurrentPage returns the content of the current page
func (b *Book) ReadCurrentPage() (string, error) {
	if !b.isOpen {
		return "", errors.New("book is closed")
	}
	return b.pages[b.currentPageIdx].GetContent(), nil
}

// HasNextPage checks if there's a next page
func (b *Book) HasNextPage() bool {
	return b.isOpen && b.currentPageIdx < len(b.pages)-1
}

// HasPrevPage checks if there's a previous page
func (b *Book) HasPrevPage() bool {
	return b.isOpen && b.currentPageIdx > 0
}
