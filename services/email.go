package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mailersend/mailersend-go"
)

const (
	// DefaultSenderEmail = "support@agrisoft-systems.com"
	DefaultSenderEmail = "MS_qxfvO4@test-dnvo4d92p5ng5r86.mlsender.net"
	DefaultSenderName  = "Thy Dungeonman"
)

type EmailToFromHeader struct {
	Name  string
	Email string
}

type Email struct {
	From      mailersend.Recipient
	To        []mailersend.Recipient
	Subject   string
	Plaintext string
	HTML      string
}

// This is currently a dummy service that just writes the URL containing the token
// to a file. Will connect it to a real SMTP server later.
type EmailService struct {
	DefaultSender    mailersend.Recipient
	MailerSendClient *mailersend.Mailersend
}

func NewEmailService(apiKey string) *EmailService {
	return &EmailService{
		MailerSendClient: mailersend.NewMailersend(apiKey),
	}
}

func (es EmailService) setFrom(msg *mailersend.Message, email Email) {
	var from mailersend.Recipient
	switch {
	case email.From != mailersend.Recipient{}:
		from = email.From
	case es.DefaultSender != mailersend.Recipient{}:
		from = es.DefaultSender
	default:
		from = mailersend.Recipient{Name: DefaultSenderName, Email: DefaultSenderEmail}
	}
	msg.SetFrom(from)
}

func (es EmailService) Send(email Email) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := es.MailerSendClient.Email.NewMessage()
	es.setFrom(message, email)
	message.SetRecipients(email.To)
	message.SetSubject(email.Subject)
	message.SetText(email.Plaintext)
	message.SetHTML(email.HTML)

	// We ignore the email response for now
	_, err := es.MailerSendClient.Email.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

// Currently only supports a single recipient
func (es EmailService) ForgotPassword(recpientName, recipientEmail, resetURL string) error {
	recipient := mailersend.Recipient{Email: recipientEmail}
	email := Email{
		Subject:   "Reset your password",
		To:        []mailersend.Recipient{recipient},
		Plaintext: "To reset your password, please visit the following link: " + resetURL,
		HTML:      `<p>To reset your password, please visit the following link: <a href="` + resetURL + `">` + resetURL + `</a></p>`,
	}
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	return nil
}
