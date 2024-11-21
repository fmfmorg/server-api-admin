package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"server-api-admin/config"
	"time"

	"github.com/julienschmidt/httprouter"
)

var Router *httprouter.Router

func init() {
	Router = httprouter.New()
}

func setSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")

		// Set HSTS header (only in production)
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		if r.Header.Get("X-Request-Source") != "SSR" {
			csrfCookie, err := r.Cookie("csrf")
			if err != nil || csrfCookie.Value != r.Header.Get("X-CSRF-Token") {
				http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func Listen(ctx context.Context) {
	handler := setSecurityHeaders(Router)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.APIPort),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", config.APIPort, err)
		}
	}()

	log.Printf("Server is ready to handle requests at %s\n", config.APIPort)

	<-ctx.Done()
	log.Println("Shutting down server...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exiting")
}

// func Listen() {
// 	handler := setSecurityHeaders(Router)
// 	server := &http.Server{
// 		Addr:         fmt.Sprintf(":%s", config.APIPort),
// 		Handler:      handler,
// 		ReadTimeout:  5 * time.Second,
// 		WriteTimeout: 10 * time.Second,
// 		IdleTimeout:  15 * time.Second,
// 	}
// 	log.Fatal(server.ListenAndServe())
// }
