package emailclient

import "waitlist/models"

// EmailClient interface
type EmailClient interface {
	Send(email *models.Message) error
}
