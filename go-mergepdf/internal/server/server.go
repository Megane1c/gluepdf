// Package server provides the HTTP server setup for go-mergepdf.
//
// NewServer creates and configures the HTTP server, session manager, and file directories.
//
// Expected outputs:
// - Server listens on the configured port (default 8080)
// - Old sessions and files are cleaned up periodically
//
// Usage:
//
//	server := server.NewServer()
//	server.ListenAndServe()
//
// See internal/server/routes.go for route registration.
package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-mergepdf/internal/session"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port           int
	SessionManager *session.SessionManager
	UploadDir      string
	OutputDir      string
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	uploadDir := "uploads"
	outputDir := "output"

	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(outputDir, 0755)

	srv := &Server{
		port:           port,
		SessionManager: session.NewSessionManager(),
		UploadDir:      uploadDir,
		OutputDir:      outputDir,
	}

	// Cleanup goroutine for old sessions/files
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			srv.SessionManager.Mutex.Lock()
			for id, session := range srv.SessionManager.Sessions {
				if time.Since(session.CreatedAt) > 5*time.Minute {
					session.Cleanup()
					delete(srv.SessionManager.Sessions, id)
				}
			}
			srv.SessionManager.Mutex.Unlock()
		}
	}()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
