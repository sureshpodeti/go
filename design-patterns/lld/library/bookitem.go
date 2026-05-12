package library

import "time"

// BookItem represents a physical copy of a book
// This is what members actually checkout (not Book)
// BookItem is responsible for: knowing its status, location, barcode
// Key Design Decision: Separate Book (metadata) from BookItem (physical copy)
type BookItem struct {
	Barcode       string
	Book          *Book      // Reference to book metadata
	Rack          *Rack      // Physical location
	Status        BookStatus
	DateOfPurchase time.Time
	Price         float64
}

func NewBookItem(barcode string, book *Book, rack *Rack, price float64) *BookItem {
	return &BookItem{
		Barcode:        barcode,
		Book:           book,
		Rack:           rack,
		Status:         StatusAvailable,
		DateOfPurchase: time.Now(),
		Price:          price,
	}
}

func (bi *BookItem) IsAvailable() bool {
	return bi.Status == StatusAvailable
}

func (bi *BookItem) Checkout() bool {
	if bi.Status != StatusAvailable {
		return false
	}
	bi.Status = StatusCheckedOut
	return true
}

func (bi *BookItem) Return() {
	bi.Status = StatusAvailable
}

func (bi *BookItem) Reserve() bool {
	if bi.Status != StatusAvailable {
		return false
	}
	bi.Status = StatusReserved
	return true
}

func (bi *BookItem) MarkLost() {
	bi.Status = StatusLost
}
