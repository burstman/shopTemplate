package app

import (
	"context"
	"log/slog"
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
	capiSvc := services.NewFacebookCAPIService()

	event.Subscribe("order.placed", func(ctx context.Context, data any) {
		slog.Debug("order.placed event triggered")
		order, ok := data.(models.Order)
		if ok {
			slog.Info("processing purchase event for Facebook CAPI", "order_id", order.ID, "total", order.Total)
			notificationSvc.NotifyAll(order)
			capiSvc.SendPurchaseEvent(order)
			slog.Info("finished processing purchase event for Facebook CAPI", "order_id", order.ID)
		}
	})

	event.Subscribe("order.abandoned", func(ctx context.Context, data any) {
		slog.Debug("order.abandoned event triggered")
		order, ok := data.(models.Order)
		if ok {
			slog.Info("processing initiate checkout event for Facebook CAPI", "order_id", order.ID)
			capiSvc.SendInitiateCheckoutEvent(order)
			notificationSvc.NotifyAbandoned(order)
		}
	})
}
