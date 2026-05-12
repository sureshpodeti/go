package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== Book Reading System Demo ===\n")

	// Create authors
	author1 := NewAuthor("auth_001", "J.K. Rowling")
	author2 := NewAuthor("auth_002", "George R.R. Martin")

	// Create a book
	harryPotter := NewBook("978-0439708180", "Harry Potter and the Sorcerer's Stone", 1997)

	// Add authors
	harryPotter.AddAuthor(author1)
	fmt.Printf("✓ Added author: %s\n", author1.GetName())

	// Add pages
	pages := []string{
		"Mr. and Mrs. Dursley, of number four, Privet Drive, were proud to say that they were perfectly normal, thank you very much.",
		"They were the last people you'd expect to be involved in anything strange or mysterious, because they just didn't hold with such nonsense.",
		"Mr. Dursley was the director of a firm called Grunnings, which made drills.",
		"He was a big, beefy man with hardly any neck, although he did have a very large mustache.",
		"Mrs. Dursley was thin and blonde and had nearly twice the usual amount of neck, which came in very useful as she spent so much of her time craning over garden fences, spying on the neighbors.",
	}

	for i, content := range pages {
		page := NewPage(i+1, content)
		harryPotter.AddPage(page)
	}
	fmt.Printf("✓ Added %d pages\n\n", len(pages))

	// Display book info
	fmt.Println("--- Book Information ---")
	fmt.Printf("Title: %s\n", harryPotter.GetTitle())
	fmt.Printf("ISBN: %s\n", harryPotter.GetISBN())
	fmt.Printf("Publication Year: %d\n", harryPotter.GetPublicationYear())
	fmt.Printf("Authors: ")
	for i, author := range harryPotter.GetAuthors() {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(author.GetName())
	}
	fmt.Printf("\nTotal Pages: %d\n\n", harryPotter.GetNumberOfPages())

	// Try to read before opening (should fail)
	fmt.Println("--- Testing Error Handling ---")
	if _, err := harryPotter.ReadCurrentPage(); err != nil {
		fmt.Printf("✓ Cannot read closed book: %v\n\n", err)
	}

	// Open the book
	fmt.Println("--- Opening the Book ---")
	if err := harryPotter.Open(); err != nil {
		fmt.Printf("Error opening book: %v\n", err)
		return
	}
	fmt.Printf("✓ Book opened successfully\n")
	fmt.Printf("✓ Current page: %d\n\n", harryPotter.GetCurrentPageNumber())

	// Read all pages sequentially
	fmt.Println("--- Reading All Pages ---")
	for {
		content, err := harryPotter.ReadCurrentPage()
		if err != nil {
			fmt.Printf("Error reading page: %v\n", err)
			break
		}

		fmt.Printf("Page %d: %s\n\n", harryPotter.GetCurrentPageNumber(), content)

		if !harryPotter.HasNextPage() {
			fmt.Println("✓ Reached the last page\n")
			break
		}

		if err := harryPotter.NextPage(); err != nil {
			fmt.Printf("Error moving to next page: %v\n", err)
			break
		}
	}

	// Navigate backwards
	fmt.Println("--- Testing Backward Navigation ---")
	if err := harryPotter.PrevPage(); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		content, _ := harryPotter.ReadCurrentPage()
		fmt.Printf("✓ Moved to previous page\nPage %d: %s\n\n", harryPotter.GetCurrentPageNumber(), content)
	}

	// Jump to specific page
	fmt.Println("--- Testing Jump to Page ---")
	if err := harryPotter.GoToPage(3); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		content, _ := harryPotter.ReadCurrentPage()
		fmt.Printf("✓ Jumped to page 3\nPage %d: %s\n\n", harryPotter.GetCurrentPageNumber(), content)
	}

	// Test boundary conditions
	fmt.Println("--- Testing Boundary Conditions ---")
	harryPotter.GoToPage(1)
	if err := harryPotter.PrevPage(); err != nil {
		fmt.Printf("✓ Cannot go before first page: %v\n", err)
	}

	harryPotter.GoToPage(harryPotter.GetNumberOfPages())
	if err := harryPotter.NextPage(); err != nil {
		fmt.Printf("✓ Cannot go beyond last page: %v\n\n", err)
	}

	// Close the book
	fmt.Println("--- Closing the Book ---")
	if err := harryPotter.Close(); err != nil {
		fmt.Printf("Error closing book: %v\n", err)
	} else {
		fmt.Println("✓ Book closed successfully")
	}

	// Try to read after closing (should fail)
	if _, err := harryPotter.ReadCurrentPage(); err != nil {
		fmt.Printf("✓ Cannot read closed book: %v\n\n", err)
	}

	// Demonstrate multiple authors
	fmt.Println("--- Testing Multiple Authors ---")
	gameOfThrones := NewBook("978-0553103540", "A Game of Thrones", 1996)
	gameOfThrones.AddAuthor(author2)
	fmt.Printf("Created book: %s by %s\n", gameOfThrones.GetTitle(), author2.GetName())

	// Demonstrate author immutability
	fmt.Println("\n--- Demonstrating Author Immutability ---")
	fmt.Printf("Author shared across books: %s\n", author1.GetName())
	fmt.Printf("Used in: %s\n", harryPotter.GetTitle())
	fmt.Println("✓ Author object is immutable - no setters available")
	fmt.Println("✓ Same author can be safely shared across multiple books")

	fmt.Println("\n=== Demo Complete ===")
}
