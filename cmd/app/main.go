package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"shopTemplate/app"
	"shopTemplate/public"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	kit.Setup()

	router := chi.NewMux()

	app.InitializeMiddleware(router)

	if kit.IsDevelopment() {
		router.Handle("/public/*", disableCache(staticDev()))
	} else if kit.IsProduction() {
		router.Handle("/public/*", staticProd())
	}

	kit.UseErrorHandler(app.ErrorHandler)
	app.InitializeRoutes(router)
	router.NotFound(kit.Handler(app.NotFoundHandler))
	app.RegisterEvents()

	listenAddr := os.Getenv("HTTP_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":3000"
	}

	// The URL to display to the user. In development, this is the Templ proxy.
	// In production, it's the application's direct address.
	displayURL := "http://localhost:7331"
	if kit.IsProduction() {
		// For production, display the actual address the server is binding to.
		// If listenAddr is ":3000", it binds to 0.0.0.0:3000, so localhost is fine for display.
		displayURL = fmt.Sprintf("http://localhost%s", listenAddr)
	}

	fmt.Printf("application running in %s\n", kit.Env())
	fmt.Printf("backend listening on: %s\n", listenAddr)
	fmt.Printf("access application via: %s\n", displayURL)

	log.Fatal(http.ListenAndServe(listenAddr, router))
}

func staticDev() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}

func staticProd() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(public.AssetsFS))
}

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func init() {
	if err := godotenv.Load(); err != nil {
		// Do not use log.Fatal here, as .env might be missing in production
		fmt.Println("Warning: .env file not found, using system environment variables")
	}
}
