package library

import (
	"fmt"
)

// Library is the main orchestrator
// Responsibility: Coordinate operations between different components
// This is where business logic lives
type Library struct {
	name                string
	catalog             *Catalog
	bookItems           map[string]*BookItem // barcode -> BookItem
	members             map[string]*Member   // memberID -> Member
	lendings            map[string]*Lending  // lendingID -> Lending
	reservations        map[string][]*Reservation // bookISBN -> queue of reservations
	notificationService *NotificationService
	lendingCounter      int
	reservationCounter  int
}

func NewLibrary(name string) *Library {
	return &Library{
		name:                name,
		catalog:             NewCatalog(),
		bookItems:           make(map[string]*BookItem),
		members:             make(map[string]*Member),
		lendings:            make(map[string]*Lending),
		reservations:        make(map[string][]*Reservation),
		notificationService: NewNotificationService(),
		lendingCounter:      1,
		reservationCounter:  1,
	}
}

// AddBook adds a book to the catalog
func (lib *Library) AddBook(book *Book) {
	lib.catalog.AddBook(book)
}

// AddBookItem adds a physical copy of a book
func (lib *Library) AddBookItem(bookItem *BookItem) {
	lib.bookItems[bookItem.Barcode] = bookItem
}

// RegisterMember adds a new member
func (lib *Library) RegisterMember(member *Member) {
	lib.members[member.ID] = member
}

// SearchBooks searches books using a strategy
func (lib *Library) SearchBooks(strategy SearchStrategy) []*Book {
	return lib.catalog.Search(strategy)
}

// CheckoutBook handles book checkout
func (lib *Library) CheckoutBook(memberID, bookBarcode string) error {
	member, exists := lib.members[memberID]
	if !exists {
		return fmt.Errorf("member not found")
	}

	if !member.IsActive() {
		return fmt.Errorf("member is not active")
	}

	// Check if member has reached checkout limit
	activeCheckouts := lib.getActiveLendingsForMember(memberID)
	if len(activeCheckouts) >= MaxBooksPerMember {
		return fmt.Errorf("member has reached maximum checkout limit (%d)", MaxBooksPerMember)
	}

	bookItem, exists := lib.bookItems[bookBarcode]
	if !exists {
		return fmt.Errorf("book item not found")
	}

	if !bookItem.Checkout() {
		return fmt.Errorf("book is not available for checkout")
	}

	// Create lending record
	lendingID := fmt.Sprintf("L%d", lib.lendingCounter)
	lib.lendingCounter++
	lending := NewLending(lendingID, bookItem, member)
	lib.lendings[lendingID] = lending

	fmt.Printf("Book '%s' checked out to %s. Due date: %s\n",
		bookItem.Book.Title, member.Name, lending.DueDate.Format("2006-01-02"))

	return nil
}

// ReturnBook handles book return
func (lib *Library) ReturnBook(bookBarcode string) error {
	bookItem, exists := lib.bookItems[bookBarcode]
	if !exists {
		return fmt.Errorf("book item not found")
	}

	// Find active lending for this book
	var activeLending *Lending
	for _, lending := range lib.lendings {
		if lending.BookItem.Barcode == bookBarcode && lending.ReturnDate == nil {
			activeLending = lending
			break
		}
	}

	if activeLending == nil {
		return fmt.Errorf("no active lending found for this book")
	}

	// Process return
	fine := activeLending.ReturnBook()
	bookItem.Return()

	if fine > 0 {
		lib.notificationService.NotifyFine(activeLending.Member, fine)
		fmt.Printf("Book returned with fine: $%.2f\n", fine)
	} else {
		fmt.Printf("Book returned successfully\n")
	}

	// Check if anyone has reserved this book
	lib.processReservations(bookItem.Book.ISBN)

	return nil
}

// ReserveBook reserves a book for a member
func (lib *Library) ReserveBook(memberID, isbn string) error {
	member, exists := lib.members[memberID]
	if !exists {
		return fmt.Errorf("member not found")
	}

	// Find the book
	books := lib.catalog.Search(&TitleSearchStrategy{Title: ""}) // Get all books
	var targetBook *Book
	for _, book := range books {
		if book.ISBN == isbn {
			targetBook = book
			break
		}
	}

	if targetBook == nil {
		return fmt.Errorf("book not found")
	}

	// Create reservation
	reservationID := fmt.Sprintf("R%d", lib.reservationCounter)
	lib.reservationCounter++
	reservation := NewReservation(reservationID, targetBook, member)

	if lib.reservations[isbn] == nil {
		lib.reservations[isbn] = make([]*Reservation, 0)
	}
	lib.reservations[isbn] = append(lib.reservations[isbn], reservation)

	fmt.Printf("Book '%s' reserved for %s\n", targetBook.Title, member.Name)
	return nil
}

// GetMemberCheckouts returns all books checked out by a member
func (lib *Library) GetMemberCheckouts(memberID string) []*Lending {
	return lib.getActiveLendingsForMember(memberID)
}

// GetBookHolder returns who has checked out a specific book
func (lib *Library) GetBookHolder(bookBarcode string) (*Member, error) {
	for _, lending := range lib.lendings {
		if lending.BookItem.Barcode == bookBarcode && lending.ReturnDate == nil {
			return lending.Member, nil
		}
	}
	return nil, fmt.Errorf("book is not checked out")
}

// Helper methods

func (lib *Library) getActiveLendingsForMember(memberID string) []*Lending {
	result := make([]*Lending, 0)
	for _, lending := range lib.lendings {
		if lending.Member.ID == memberID && lending.ReturnDate == nil {
			result = append(result, lending)
		}
	}
	return result
}

func (lib *Library) processReservations(isbn string) {
	reservations, exists := lib.reservations[isbn]
	if !exists || len(reservations) == 0 {
		return
	}

	// Find first pending reservation
	for i, reservation := range reservations {
		if reservation.IsPending() {
			lib.notificationService.NotifyBookAvailable(reservation.Member, reservation.Book)
			reservation.Fulfill()
			// Remove from queue
			lib.reservations[isbn] = append(reservations[:i], reservations[i+1:]...)
			break
		}
	}
}
