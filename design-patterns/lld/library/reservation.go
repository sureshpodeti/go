package library

import "time"

// Reservation represents a member's reservation for a book
// When a book becomes available, the first reservation gets notified
type Reservation struct {
	ID              string
	Book            *Book
	Member          *Member
	ReservationDate time.Time
	Status          string // PENDING, FULFILLED, CANCELLED
}

func NewReservation(id string, book *Book, member *Member) *Reservation {
	return &Reservation{
		ID:              id,
		Book:            book,
		Member:          member,
		ReservationDate: time.Now(),
		Status:          "PENDING",
	}
}

func (r *Reservation) Fulfill() {
	r.Status = "FULFILLED"
}

func (r *Reservation) Cancel() {
	r.Status = "CANCELLED"
}

func (r *Reservation) IsPending() bool {
	return r.Status == "PENDING"
}
