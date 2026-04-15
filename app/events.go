package app

import (
	"context"
	"shopTemplate/app/models"
	"shopTemplate/app/services"

	"github.com/anthdm/superkit/event"
)

// RegisterEvents configures the global event listeners for the application.
func RegisterEvents() {
	// Initialize the notification service and register providers.
	notificationSvc := services.NewNotificationService()
	notificationSvc.Register(services.NewEmailNotifier())
	notificationSvc.Register(services.NewTelegramNotifier())

	// Listen for the order.placed event emitted during checkout.
	event.Subscribe("order.placed", func(ctx context.Context, data any) {
		order, ok := data.(models.Order)
		if ok {
			notificationSvc.NotifyAll(order)
		}
	})
}
