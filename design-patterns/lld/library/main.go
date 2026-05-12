package main

import (
	"fmt"
	"time"

	lib "github.com/yourusername/design-patterns/lld/library"
)

func main() {
	fmt.Println("=== Library Management System Demo ===\n")

	// Initialize library
	library := lib.NewLibrary("City Central Library")

	// Create authors
	author1 := lib.NewAuthor("A1", "Robert Martin", "USA")
	author2 := lib.NewAuthor("A2", "Martin Fowler", "UK")

	// Create books
	book1 := lib.NewBook("978-0132350884", "Clean Code", "Software Engineering",
		time.Date(2008, 8, 1, 0, 0, 0, 0, time.UTC), 464)
	book1.AddAuthor(author1)

	book2 := lib.NewBook("978-0201633610", "Design Patterns", "Software Engineering",
		time.Date(1994, 10, 31, 0, 0, 0, 0, time.UTC), 395)
	book2.AddAuthor(author2)

	// Add books to library
	library.AddBook(book1)
	library.AddBook(book2)

	// Create racks
	rack1 := lib.NewRack("A-101", "Floor 1, Section A")
	rack2 := lib.NewRack("A-102", "Floor 1, Section A")

	// Create physical book items (copies)
	bookItem1 := lib.NewBookItem("BC001", book1, rack1, 45.99)
	bookItem2 := lib.NewBookItem("BC002", book1, rack1, 45.99) // Another copy
	bookItem3 := lib.NewBookItem("BC003", book2, rack2, 54.99)

	library.AddBookItem(bookItem1)
	library.AddBookItem(bookItem2)
	library.AddBookItem(bookItem3)

	// Register members
	member1 := lib.NewMember("M001", "Alice Johnson", "alice@email.com", "555-0101", "CARD001")
	member2 := lib.NewMember("M002", "Bob Smith", "bob@email.com", "555-0102", "CARD002")

	library.RegisterMember(member1)
	library.RegisterMember(member2)

	// Demo 1: Search books by title
	fmt.Println("--- Search by Title: 'Clean' ---")
	results := library.SearchBooks(&lib.TitleSearchStrategy{Title: "Clean"})
	for _, book := range results {
		fmt.Printf("Found: %s by %s\n", book.Title, book.GetAuthorNames())
	}
	fmt.Println()

	// Demo 2: Search by author
	fmt.Println("--- Search by Author: 'Martin' ---")
	results = library.SearchBooks(&lib.AuthorSearchStrategy{AuthorName: "Martin"})
	for _, book := range results {
		fmt.Printf("Found: %s by %s\n", book.Title, book.GetAuthorNames())
	}
	fmt.Println()

	// Demo 3: Checkout book
	fmt.Println("--- Alice checks out Clean Code (copy 1) ---")
	err := library.CheckoutBook("M001", "BC001")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Demo 4: Try to checkout same book again (should fail)
	fmt.Println("--- Bob tries to checkout same copy (should fail) ---")
	err = library.CheckoutBook("M002", "BC001")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Demo 5: Checkout another copy (should succeed)
	fmt.Println("--- Bob checks out Clean Code (copy 2) ---")
	err = library.CheckoutBook("M002", "BC002")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Demo 6: Get member's checkouts
	fmt.Println("--- Alice's current checkouts ---")
	checkouts := library.GetMemberCheckouts("M001")
	for _, lending := range checkouts {
		fmt.Printf("- %s (Due: %s)\n",
			lending.BookItem.Book.Title,
			lending.DueDate.Format("2006-01-02"))
	}
	fmt.Println()

	// Demo 7: Find who has a book
	fmt.Println("--- Who has book BC001? ---")
	holder, err := library.GetBookHolder("BC001")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%s has this book\n", holder.Name)
	}
	fmt.Println()

	// Demo 8: Reserve a book
	fmt.Println("--- Alice reserves Design Patterns ---")
	err = library.ReserveBook("M001", "978-0201633610")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Demo 9: Return book (on time)
	fmt.Println("--- Alice returns Clean Code ---")
	err = library.ReturnBook("BC001")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Demo 10: Simulate overdue return
	fmt.Println("--- Simulating overdue scenario ---")
	member3 := lib.NewMember("M003", "Charlie Brown", "charlie@email.com", "555-0103", "CARD003")
	library.RegisterMember(member3)

	bookItem4 := lib.NewBookItem("BC004", book2, rack2, 54.99)
	library.AddBookItem(bookItem4)

	// Checkout and manually set due date to past
	library.CheckoutBook("M003", "BC004")
	// In real scenario, this would be overdue after 10 days
	fmt.Println("(In production, fine would be calculated after 10 days)")
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}
