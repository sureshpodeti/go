# Library Management System

A comprehensive low-level design implementation of a library management system in Go.

## Features Implemented

✅ Search books by title, author, subject, publication date  
✅ Unique identification (ISBN for books, barcode for copies)  
✅ Multiple copies of same book (Book vs BookItem)  
✅ Checkout and return functionality  
✅ Track who has which book  
✅ Maximum 5 books per member limit  
✅ Maximum 10 days lending period  
✅ Fine calculation for overdue books  
✅ Book reservation system  
✅ Notifications for available books and fines  
✅ Barcode support for books and member cards  

## Architecture

### Core Design Decisions

1. **Book vs BookItem Separation**
   - `Book`: Metadata (title, ISBN, authors) - one record
   - `BookItem`: Physical copy with barcode, rack location - multiple records
   - Why? One book can have multiple physical copies

2. **Search Strategy Pattern**
   - Different search strategies: Title, Author, Subject, Publication Date
   - Easy to add new search criteria without modifying existing code

3. **Responsibility Distribution**
   - `Library`: Orchestrates all operations
   - `Catalog`: Handles book searching
   - `Lending`: Tracks checkout/return and calculates fines
   - `NotificationService`: Sends notifications
   - `Member`: Knows personal information
   - `BookItem`: Manages availability status

## Project Structure

```
lld/library/
├── author.go           # Author entity
├── book.go             # Book metadata
├── bookitem.go         # Physical book copy
├── constants.go        # System constants and enums
├── lending.go          # Checkout transaction
├── library.go          # Main orchestrator
├── member.go           # Library member
├── notification.go     # Notification service
├── rack.go             # Physical location
├── reservation.go      # Book reservation
├── search.go           # Search strategies
├── main.go             # Demo application
├── docs/
│   ├── DESIGN_APPROACH.md  # Step-by-step design thinking
│   └── GOLDEN_RULES.md     # LLD problem-solving guide
└── README.md
```

## Running the Demo

```bash
cd lld/library
go run main.go
```

## Key Classes

### Book
Represents book metadata (abstract concept)
- ISBN, title, authors, subject, publication date
- One book can have multiple physical copies

### BookItem
Represents a physical copy of a book
- Barcode, rack location, status, price
- This is what members actually checkout

### Member
Represents a library member
- ID, name, email, card barcode, status
- Can checkout up to 5 books

### Library
Main orchestrator that coordinates:
- Book catalog and searching
- Checkout/return operations
- Reservation management
- Member management

### Lending
Tracks a checkout transaction
- Checkout date, due date, return date
- Calculates fines for overdue returns

## Design Patterns Used

- **Strategy Pattern**: Search strategies
- **Service Layer**: NotificationService
- **Repository Pattern**: Catalog for book management
- **Value Objects**: Status enums, constants

## Configuration

```go
const (
    MaxBooksPerMember = 5
    MaxLendingDays    = 10
    FinePerDay        = 1.0 // dollars
)
```

## Example Usage

```go
// Create library
library := NewLibrary("City Library")

// Add book
book := NewBook("978-0132350884", "Clean Code", "Software", pubDate, 464)
library.AddBook(book)

// Add physical copy
rack := NewRack("A-101", "Floor 1")
bookItem := NewBookItem("BC001", book, rack, 45.99)
library.AddBookItem(bookItem)

// Register member
member := NewMember("M001", "Alice", "alice@email.com", "555-0101", "CARD001")
library.RegisterMember(member)

// Search books
results := library.SearchBooks(&TitleSearchStrategy{Title: "Clean"})

// Checkout
library.CheckoutBook("M001", "BC001")

// Return
library.ReturnBook("BC001")
```

## Learning Resources

- `docs/DESIGN_APPROACH.md`: Detailed explanation of design decisions
- `docs/GOLDEN_RULES.md`: Step-by-step guide to solving LLD problems

## Future Enhancements

- Database persistence
- REST API endpoints
- Email/SMS integration for notifications
- Payment gateway for fines
- Book renewal functionality
- Librarian role with admin privileges
