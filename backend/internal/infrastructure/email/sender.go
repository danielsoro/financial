package email

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Sender is the interface for sending emails.
type Sender interface {
	Send(to, subject, htmlBody string) error
}

// SendGridSender sends emails via SendGrid API.
type SendGridSender struct {
	apiKey string
	from   string
}

func NewSender(apiKey, from string) Sender {
	if apiKey == "" {
		return &LogSender{}
	}
	return &SendGridSender{apiKey: apiKey, from: from}
}

func (s *SendGridSender) Send(to, subject, htmlBody string) error {
	from := mail.NewEmail("", s.from)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, "", htmlBody)
	client := sendgrid.NewSendClient(s.apiKey)
	resp, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("sendgrid: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("sendgrid: status %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

// LogSender logs emails to stdout instead of sending them. Used in development.
type LogSender struct{}

func (s *LogSender) Send(to, subject, htmlBody string) error {
	log.Printf("[EMAIL] To: %s | Subject: %s", to, subject)
	return nil
}
