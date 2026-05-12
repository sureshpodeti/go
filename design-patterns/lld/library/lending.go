package library

import "time"

// Lending represents a checkout transaction
// Lending is responsible for: tracking checkout/return dates, calculating fines
type Lending struct {
	ID           string
	BookItem     *BookItem
	Member       *Member
	CheckoutDate time.Time
	DueDate      time.Time
	ReturnDate   *time.Time // nil if not returned yet
	Fine         float64
}

func NewLending(id string, bookItem *BookItem, member *Member) *Lending {
	checkoutDate := time.Now()
	return &Lending{
		ID:           id,
		BookItem:     bookItem,
		Member:       member,
		CheckoutDate: checkoutDate,
		DueDate:      CalculateDueDate(checkoutDate),
		ReturnDate:   nil,
		Fine:         0,
	}
}

func (l *Lending) IsOverdue() bool {
	if l.ReturnDate != nil {
		return false // Already returned
	}
	return time.Now().After(l.DueDate)
}

func (l *Lending) CalculateFine() float64 {
	if l.ReturnDate == nil {
		// Not yet returned, calculate based on current date
		if time.Now().After(l.DueDate) {
			daysOverdue := int(time.Since(l.DueDate).Hours() / 24)
			return float64(daysOverdue) * FinePerDay
		}
		return 0
	}
	
	// Already returned, calculate based on return date
	if l.ReturnDate.After(l.DueDate) {
		daysOverdue := int(l.ReturnDate.Sub(l.DueDate).Hours() / 24)
		return float64(daysOverdue) * FinePerDay
	}
	return 0
}

func (l *Lending) ReturnBook() float64 {
	now := time.Now()
	l.ReturnDate = &now
	l.Fine = l.CalculateFine()
	return l.Fine
}
