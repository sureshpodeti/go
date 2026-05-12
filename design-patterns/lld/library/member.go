package library

import "time"

// Member represents a library member
// Member is responsible for: knowing personal info, tracking their checkouts
// Member is NOT responsible for: searching books, calculating fines
type Member struct {
	ID              string
	Name            string
	Email           string
	Phone           string
	CardBarcode     string // Unique barcode on member card
	Status          MemberStatus
	DateOfMembership time.Time
}

func NewMember(id, name, email, phone, cardBarcode string) *Member {
	return &Member{
		ID:               id,
		Name:             name,
		Email:            email,
		Phone:            phone,
		CardBarcode:      cardBarcode,
		Status:           MemberActive,
		DateOfMembership: time.Now(),
	}
}

func (m *Member) IsActive() bool {
	return m.Status == MemberActive
}

func (m *Member) Block() {
	m.Status = MemberBlocked
}

func (m *Member) Activate() {
	m.Status = MemberActive
}
