package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"shopTemplate/app/config"
	"shopTemplate/app/models"
)

// TelegramNotifier implements the OrderNotifier interface using Telegram Bot API.
type TelegramNotifier struct{}

// NewTelegramNotifier creates a new instance.
func NewTelegramNotifier() *TelegramNotifier {
	return &TelegramNotifier{}
}

func (t *TelegramNotifier) Name() string {
	return "Telegram Bot"
}

// Send sends the order notification to the configured Telegram chat.
func (t *TelegramNotifier) Send(order models.Order) error {
	cfg := config.Get()
	token := cfg.Notification.TelegramBotToken
	chatID := cfg.Notification.TelegramChatID

	// Skip if not configured
	if token == "" || chatID == "" {
		return nil
	}

	text := fmt.Sprintf(
		"🔔 *New Order Received #%d*\n\n"+
			"*Customer:* %s %s (%s)\n"+
			"*Total Amount:* $%.2f\n"+
			"*Shipping Address:* %s, %s\n"+
			"*Customer Phone:* %s\n\n"+
			"Please log in to your admin panel to manage this order.",
		order.ID, order.FirstName, order.LastName, order.Email, order.Total, order.Address, order.City, order.Phone,
	)

	return t.sendMessage(token, chatID, text)
}

// SendTest sends a simple verification message to the specified bot and chat.
func (t *TelegramNotifier) SendTest(token, chatID string) error {
	text := "🚀 This is a test notification from your shop. Your Telegram configuration is correct!"
	return t.sendMessage(token, chatID, text)
}

func (t *TelegramNotifier) sendMessage(token, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %d", resp.StatusCode)
	}

	return nil
}
