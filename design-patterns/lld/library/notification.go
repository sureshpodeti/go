package library

import "fmt"

// NotificationService handles sending notifications to members
// Single Responsibility: Only notification logic
type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (ns *NotificationService) NotifyBookAvailable(member *Member, book *Book) {
	// In real system, this would send email/SMS
	fmt.Printf("[NOTIFICATION] Dear %s, the book '%s' is now available for checkout.\n",
		member.Name, book.Title)
}

func (ns *NotificationService) NotifyOverdue(member *Member, lending *Lending) {
	// In real system, this would send email/SMS
	fmt.Printf("[NOTIFICATION] Dear %s, the book '%s' is overdue. Please return it soon.\n",
		member.Name, lending.BookItem.Book.Title)
}

func (ns *NotificationService) NotifyFine(member *Member, fine float64) {
	// In real system, this would send email/SMS
	fmt.Printf("[NOTIFICATION] Dear %s, you have a fine of $%.2f. Please pay at the counter.\n",
		member.Name, fine)
}
