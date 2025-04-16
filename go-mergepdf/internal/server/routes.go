// Package server sets up the HTTP server and registers API routes for go-mergepdf.
//
// RegisterRoutes returns an http.Handler with all API endpoints for session and PDF management.
//
// Expected outputs:
// - All API endpoints are available under /api/sessions
// - CORS and logging middleware are enabled
//
// See README.md for endpoint details and integration examples.
package server

import (
	"net"
	"net/http"

	_ "go-mergepdf/docs"
	"go-mergepdf/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Only allow requests from localhost to /swagger/*
func localhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if host != "127.0.0.1" && host != "::1" && host != "localhost" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	r.With(localhostOnly).Get("/swagger/*", httpSwagger.WrapHandler)
	h := handlers.NewAPIHandler(s.SessionManager, s.UploadDir, s.OutputDir)
	r.Route("/api/sessions", func(api chi.Router) {
		api.Post("/", h.CreateSession)
		api.Post("/{sessionID}/files", h.UploadFile)
		api.Post("/{sessionID}/signature", h.UploadSignature)
		api.Put("/{sessionID}/order", h.UpdateOrder)
		api.Post("/{sessionID}/actions/merge", h.MergeFiles)
		api.Post("/{sessionID}/sign", h.SignPDF)
		api.Get("/{sessionID}/files/{filename}", h.DownloadFile)
	})

	return r
}
