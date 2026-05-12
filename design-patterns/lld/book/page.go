package main

// Page represents a single page in a book
type Page struct {
	number  int
	content string
}

// NewPage creates a new page
func NewPage(number int, content string) *Page {
	return &Page{
		number:  number,
		content: content,
	}
}

// GetPageNo returns the page number
func (p *Page) GetPageNo() int {
	return p.number
}

// GetContent returns the page content
func (p *Page) GetContent() string {
	return p.content
}
