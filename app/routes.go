package app

import (
	"log/slog"
	"shopTemplate/app/handlers"
	"shopTemplate/app/views/errors"
	"shopTemplate/plugins/auth"

	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/kit/middleware"
	"github.com/go-chi/chi/v5"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Define your global middleware
func InitializeMiddleware(router *chi.Mux) {
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.WithRequest)
}

// Define your routes in here
func InitializeRoutes(router *chi.Mux) {
	// Authentication plugin
	//
	// By default the auth plugin is active, to disable the auth plugin
	// you will need to pass your own handler in the `AuthFunc`` field
	// of the `kit.AuthenticationConfig`.
	//  authConfig := kit.AuthenticationConfig{
	//      AuthFunc: YourAuthHandler,
	//      RedirectURL: "/login",
	//  }
	auth.InitializeRoutes(router)
	authConfig := kit.AuthenticationConfig{
		AuthFunc:    handlers.HandleAuthentication,
		RedirectURL: "/login",
	}

	// WebSocket route without the standard HTTP logger middleware to avoid duration noise
	router.With(kit.WithAuthentication(authConfig, false)).Get("/api/chat/ws", kit.Handler(handlers.HandleChatWS))

	// Routes that "might" have an authenticated user
	router.Group(func(app chi.Router) {
		app.Use(chimiddleware.Logger)
		app.Use(kit.WithAuthentication(authConfig, false)) // strict set to false

		// Routes
		app.Get("/", kit.Handler(handlers.HandleLandingIndex))
		app.Get("/products", kit.Handler(handlers.HandleProductsIndex))
		app.Get("/health", kit.Handler(handlers.HandleHealthCheck)) // Health check endpoint
		app.Get("/products/{id}", kit.Handler(handlers.HandleProductShow))
		app.Get("/categories/{id}", kit.Handler(handlers.HandleCategoryShow))
		app.Get("/products/{id}/quick-view", kit.Handler(handlers.HandleProductQuickView))
		app.Post("/cart/add/{id}", kit.Handler(handlers.HandleCartAdd))
		app.Delete("/cart/remove/{id}", kit.Handler(handlers.HandleCartRemove))
		app.Get("/cart", kit.Handler(handlers.HandleCartShow))
		app.Get("/checkout", kit.Handler(handlers.HandleCheckoutIndex))
		app.Get("/checkout/success", kit.Handler(handlers.HandleCheckoutSuccess))
		app.Post("/checkout", kit.Handler(handlers.HandleCheckoutCreate))
		app.Get("/api/chat/messages", kit.Handler(handlers.HandleChatFetchMessages))
		app.Post("/api/chat/send", kit.Handler(handlers.HandleChatSend))
	})

	// Authenticated routes
	//
	// Routes that "must" have an authenticated user or else they
	// will be redirected to the configured redirectURL, set in the
	// AuthenticationConfig.
	router.Group(func(app chi.Router) {
		app.Use(chimiddleware.Logger)
		app.Use(kit.WithAuthentication(authConfig, true)) // strict set to true

		// Routes
		app.Get("/admin/categories", kit.Handler(handlers.HandleAdminCategoriesIndex))
		app.Post("/admin/categories", kit.Handler(handlers.HandleAdminCategoryCreate))
		app.Delete("/admin/categories/{id}", kit.Handler(handlers.HandleAdminCategoryDelete))
		app.Post("/admin/categories/reorder", kit.Handler(handlers.HandleAdminCategoryReorder))
		app.Get("/admin/orders", kit.Handler(handlers.HandleAdminOrdersIndex))
		app.Get("/admin/orders/{id}", kit.Handler(handlers.HandleAdminOrderShow))
		app.Post("/admin/orders/{id}/status", kit.Handler(handlers.HandleAdminOrderUpdateStatus))
		app.Get("/admin/orders/{id}/delete", kit.Handler(handlers.HandleAdminOrderDeleteConfirm))
		app.Delete("/admin/orders/{id}", kit.Handler(handlers.HandleAdminOrderDelete))
		app.Get("/admin/products", kit.Handler(handlers.HandleAdminProductsIndex))
		app.Get("/admin/users", kit.Handler(handlers.HandleAdminUsersIndex))
		app.Get("/admin/users/{id}/edit", kit.Handler(handlers.HandleAdminUserEdit))
		app.Put("/admin/users/{id}", kit.Handler(handlers.HandleAdminUserUpdate))
		app.Get("/admin/chats", kit.Handler(handlers.HandleAdminChatIndex))
		app.Get("/admin/chat/{id}", kit.Handler(handlers.HandleAdminChatShow))
		app.Post("/admin/chat/{id}/send", kit.Handler(handlers.HandleAdminChatSend))
		app.Post("/admin/chat/{id}/ban", kit.Handler(handlers.HandleAdminChatBan))
		app.Get("/admin/chats/sidebar", kit.Handler(handlers.HandleAdminChatSidebar))       // New handler for sidebar polling
		app.Get("/admin/chat/{id}/messages", kit.Handler(handlers.HandleAdminChatMessages)) // New handler for message polling

		app.Get("/configuration", kit.Handler(handlers.HandleConfigurationIndex))
		app.Get("/admin/{section}", kit.Handler(handlers.HandleAdminSettings))
		app.Post("/admin/{section}", kit.Handler(handlers.HandleAdminSettingsUpdate))
		app.Post("/admin/notifications/test", kit.Handler(handlers.HandleAdminNotificationsTest))
		app.Post("/admin/notifications/test/telegram", kit.Handler(handlers.HandleAdminTelegramNotificationsTest))
		app.Post("/admin/sections/add", kit.Handler(handlers.HandleAdminSectionAdd))
		app.Post("/admin/sections/{index}/delete", kit.Handler(handlers.HandleAdminSectionDelete))
		app.Post("/admin/sections/{index}/duplicate", kit.Handler(handlers.HandleAdminSectionDuplicate))
		app.Get("/products/new", kit.Handler(handlers.HandleProductNew))
		app.Post("/products", kit.Handler(handlers.HandleProductCreate))
		app.Get("/products/{id}/edit", kit.Handler(handlers.HandleProductEdit))
		app.Put("/products/{id}", kit.Handler(handlers.HandleProductUpdate))
		app.Get("/products/{id}/delete", kit.Handler(handlers.HandleProductDeleteConfirm))
		app.Delete("/products/{id}", kit.Handler(handlers.HandleProductDelete))
		// app.Get("/path", kit.Handler(myHandler.HandleIndex))
	})
}

// NotFoundHandler that will be called when the requested path could
// not be found.
func NotFoundHandler(kit *kit.Kit) error {
	// Silence warnings for templ live-reload endpoint during development
	if kit.Request.URL.Path == "/_templ/reload" {
		return nil
	}

	slog.Warn("not found", "path", kit.Request.URL.Path)
	return kit.Render(errors.Error404())
}

// ErrorHandler that will be called on errors return from application handlers.
func ErrorHandler(kit *kit.Kit, err error) {
	slog.Error("internal server error", "err", err.Error(), "path", kit.Request.URL.Path)
	kit.Render(errors.Error500())
}
