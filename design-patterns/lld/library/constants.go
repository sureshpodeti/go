package library

import "time"

// System configuration constants
const (
	MaxBooksPerMember = 5
	MaxLendingDays    = 10
	FinePerDay        = 1.0 // dollars
)

// BookStatus represents the status of a book item
type BookStatus string

const (
	StatusAvailable  BookStatus = "AVAILABLE"
	StatusCheckedOut BookStatus = "CHECKED_OUT"
	StatusReserved   BookStatus = "RESERVED"
	StatusLost       BookStatus = "LOST"
)

// MemberStatus represents the status of a library member
type MemberStatus string

const (
	MemberActive   MemberStatus = "ACTIVE"
	MemberInactive MemberStatus = "INACTIVE"
	MemberBlocked  MemberStatus = "BLOCKED"
)

// Helper function to calculate due date
func CalculateDueDate(checkoutDate time.Time) time.Time {
	return checkoutDate.AddDate(0, 0, MaxLendingDays)
}
