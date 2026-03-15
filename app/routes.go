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
	router.Use(chimiddleware.Logger)
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
		AuthFunc:    auth.AuthenticateUser,
		RedirectURL: "/login",
	}

	// Routes that "might" have an authenticated user
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, false)) // strict set to false

		// Routes
		app.Get("/", kit.Handler(handlers.HandleLandingIndex))
		app.Get("/plants", kit.Handler(handlers.HandlePlantsIndex))
	})

	// Authenticated routes
	//
	// Routes that "must" have an authenticated user or else they
	// will be redirected to the configured redirectURL, set in the
	// AuthenticationConfig.
	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, true)) // strict set to true

		// Routes
		app.Get("/configuration", kit.Handler(handlers.HandleConfigurationIndex))
		app.Post("/configuration", kit.Handler(handlers.HandleConfigurationUpdate))
		app.Get("/products/new", kit.Handler(handlers.HandleProductNew))
		app.Post("/products", kit.Handler(handlers.HandleProductCreate))
		app.Get("/products/{id}/edit", kit.Handler(handlers.HandleProductEdit))
		app.Put("/products/{id}", kit.Handler(handlers.HandleProductUpdate))
		app.Delete("/products/{id}", kit.Handler(handlers.HandleProductDelete))
		// app.Get("/path", kit.Handler(myHandler.HandleIndex))
	})
}

// NotFoundHandler that will be called when the requested path could
// not be found.
func NotFoundHandler(kit *kit.Kit) error {
	return kit.Render(errors.Error404())
}

// ErrorHandler that will be called on errors return from application handlers.
func ErrorHandler(kit *kit.Kit, err error) {
	slog.Error("internal server error", "err", err.Error(), "path", kit.Request.URL.Path)
	kit.Render(errors.Error500())
}
