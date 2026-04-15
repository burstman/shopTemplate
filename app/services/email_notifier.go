package services

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"shopTemplate/app/config"
	"shopTemplate/app/models"
)

// EmailNotifier implements the OrderNotifier interface using SMTP.
type EmailNotifier struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewEmailNotifier creates a new instance using environment variables.
func NewEmailNotifier() *EmailNotifier {
	if os.Getenv("SMTP_PASS") == "" {
		slog.Warn("SMTP_PASS is not set in environment variables")
	}

	return &EmailNotifier{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASS"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (e *EmailNotifier) Name() string {
	return "Brevo SMTP"
}

// Send formats a basic MIME email and sends it via the configured SMTP server.
func (e *EmailNotifier) Send(order models.Order) error {
	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	cfg := config.Get()

	adminEmail := cfg.Notification.AdminEmailRecipient
	if adminEmail == "" {
		adminEmail = e.from
	}

	// 1. Send Confirmation to Customer
	customerHeader := fmt.Sprintf("From: %s\r\n", e.from)
	customerHeader += fmt.Sprintf("To: %s\r\n", order.Email)
	customerHeader += fmt.Sprintf("Subject: Order Confirmation #%d - %s\r\n", order.ID, cfg.Site.Name)
	customerHeader += "MIME-version: 1.0\r\n"
	customerHeader += "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	customerBody := fmt.Sprintf(
		"Hello %s %s,\n\n"+
			"Thank you for your order! We have received it and are currently processing it.\n\n"+
			"Order Summary:\n"+
			"Order Number: #%d\n"+
			"Total: $%.2f\n"+
			"Shipping to: %s, %s\n\n"+
			"We will notify you once your order has been shipped.",
		order.FirstName, order.LastName, order.ID, order.Total, order.Address, order.City,
	)

	if err := smtp.SendMail(e.host+":"+e.port, auth, e.from, []string{order.Email}, []byte(customerHeader+customerBody)); err != nil {
		slog.Error("failed to send customer confirmation email", "err", err, "orderID", order.ID)
	}

	// 2. Send Notification to Admin
	adminHeader := fmt.Sprintf("From: %s\r\n", e.from)
	adminHeader += fmt.Sprintf("To: %s\r\n", adminEmail)
	adminHeader += fmt.Sprintf("Subject: New Order #%d from %s %s\r\n", order.ID, order.FirstName, order.LastName)
	adminHeader += "MIME-version: 1.0\r\n"
	adminHeader += "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	adminBody := fmt.Sprintf(
		"Hello,\n\n"+
			"You have a new order from %s %s (%s).\n\n"+
			"Order Details:\n"+
			"Order ID: #%d\n"+
			"Total Amount: $%.2f\n"+
			"Shipping Address: %s, %s\n"+
			"Customer Phone: %s\n\n"+
			"Please log in to your admin panel to manage this order.",
		order.FirstName, order.LastName, order.Email, order.ID, order.Total, order.Address, order.City, order.Phone,
	)

	return smtp.SendMail(e.host+":"+e.port, auth, e.from, []string{adminEmail}, []byte(adminHeader+adminBody))
}

// SendTest sends a simple verification email to the specified recipient.
func (e *EmailNotifier) SendTest(recipient string) error {
	auth := smtp.PlainAuth("", e.username, e.password, e.host)

	header := fmt.Sprintf("From: %s\r\n", e.from)
	header += fmt.Sprintf("To: %s\r\n", recipient)
	header += "Subject: [Test] Shop Notification System\r\n"
	header += "MIME-version: 1.0\r\n"
	header += "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	body := "This is a test email to verify your shop's notification settings. If you received this, your SMTP configuration is working correctly."

	msg := []byte(header + body)

	return smtp.SendMail(e.host+":"+e.port, auth, e.from, []string{recipient}, msg)
}
