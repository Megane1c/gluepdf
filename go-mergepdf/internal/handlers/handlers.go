// Package handlers provides HTTP handlers for the PDF merging API.
//
// This package contains the main HTTP endpoints for session management,
// file upload, file ordering, PDF merging, and download.
//
// Example usage:
//
//	h := handlers.NewAPIHandler(sessionManager, uploadDir, outputDir)
//	r := chi.NewRouter()
//	r.Post("/api/sessions/", h.CreateSession)
//
// All handlers are designed to be used with the chi router.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"go-mergepdf/internal/pdf"
	"go-mergepdf/internal/session"
	"go-mergepdf/internal/utils"

	"github.com/go-chi/chi/v5"
)

type APIHandler struct {
	SessionManager *session.SessionManager
	UploadDir      string
	OutputDir      string
}

func NewAPIHandler(sm *session.SessionManager, uploadDir, outputDir string) *APIHandler {
	return &APIHandler{SessionManager: sm, UploadDir: uploadDir, OutputDir: outputDir}
}

// CreateSession godoc
// @Summary      Create a new session
// @Description  Creates a new PDF merge session and returns a session ID
// @Tags         sessions
// @Produce      json
// @Success      200  {object}  map[string]string  "{ sessionId: string }"
// @Router       /api/sessions/ [post]
func (h *APIHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	session := h.SessionManager.CreateSession()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"sessionId": "%s"}`, session.ID)
}

// UploadFile godoc
// @Summary      Upload a PDF file
// @Description  Uploads a PDF file to the session
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        sessionID  path      string  true  "Session ID"
// @Param        pdf        formData  file    true  "PDF file"
// @Success      200  {object}  map[string]interface{}  "{ filename: string, size: int }"
// @Failure      400  {string}  string  "Bad request"
// @Failure      404  {string}  string  "Session not found"
// @Router       /api/sessions/{sessionID}/files [post]
func (h *APIHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	session, exists := h.SessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	const maxUploadSize = 25 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sanitizeFilename := utils.SanitizeFilename(handler.Filename)
	if filepath.Ext(handler.Filename) != ".pdf" {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	header := make([]byte, 5)
	if _, err := file.Read(header); err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	if string(header) != "%PDF-" {
		http.Error(w, "Uploaded file is not a valid PDF", http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s-%s", utils.GenerateUUID(), sanitizeFilename)
	filepath := filepath.Join(h.UploadDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	session.AddFile(filepath)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"filename": "%s", "size": %d}`, filepath, handler.Size)
}

// UpdateOrder godoc
// @Summary      Set file order
// @Description  Sets the order of uploaded files for merging
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "Session ID"
// @Param        files      body      object  true  "{ files: [string] }"
// @Success      200  {object}  map[string]bool  "{ success: true }"
// @Failure      400  {string}  string  "Bad request"
// @Failure      404  {string}  string  "Session not found"
// @Router       /api/sessions/{sessionID}/order [put]
func (h *APIHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	session, exists := h.SessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	var fileOrder struct {
		Files []string `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&fileOrder); err != nil {
		http.Error(w, "Invalid file order data", http.StatusBadRequest)
		return
	}
	currentFiles := session.GetFiles()
	fileMap := make(map[string]bool)
	for _, file := range currentFiles {
		fileMap[file] = true
	}
	for _, file := range fileOrder.Files {
		if !fileMap[file] {
			http.Error(w, "Invalid file in order list", http.StatusBadRequest)
			return
		}
	}
	if len(fileOrder.Files) > 0 {
		session.SetFiles(fileOrder.Files)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true}`)
}

// MergeFiles godoc
// @Summary      Merge uploaded files
// @Description  Merges all uploaded files in the session and returns a download URL
// @Tags         files
// @Produce      json
// @Param        sessionID  path      string  true  "Session ID"
// @Success      200  {object}  map[string]string  "{ downloadUrl: string }"
// @Failure      400  {string}  string  "No files to merge"
// @Failure      404  {string}  string  "Session not found"
// @Failure      409  {string}  string  "Merge already in progress or done"
// @Router       /api/sessions/{sessionID}/actions/merge [post]
func (h *APIHandler) MergeFiles(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	session, exists := h.SessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.Mutex.Lock()
	if session.MergeStatus == "in_progress" {
		session.Mutex.Unlock()
		http.Error(w, "Merge already in progress", http.StatusConflict)
		return
	}
	if session.MergeStatus == "done" {
		session.Mutex.Unlock()
		http.Error(w, "Files already merged", http.StatusConflict)
		return
	}
	session.MergeStatus = "in_progress"
	session.Mutex.Unlock()

	files := session.GetFiles()
	if len(files) == 0 {
		session.Mutex.Lock()
		session.MergeStatus = "idle"
		session.Mutex.Unlock()
		http.Error(w, "No files to merge", http.StatusBadRequest)
		return
	}

	outputFilename := fmt.Sprintf("merged-%s.pdf", utils.GenerateUUID())
	outputPath := filepath.Join(h.OutputDir, outputFilename)
	if err := pdf.MergePDFs(files, outputPath); err != nil {
		session.Mutex.Lock()
		session.MergeStatus = "idle"
		session.Mutex.Unlock()
		log.Printf("Error merging PDFs: %v", err)
		http.Error(w, "Failed to merge PDFs", http.StatusInternalServerError)
		return
	}
	if err := pdf.RemoveBookmarks(outputPath); err != nil {
		session.Mutex.Lock()
		session.MergeStatus = "idle"
		session.Mutex.Unlock()
		log.Printf("Error removing bookmarks: %v", err)
		http.Error(w, "Failed to process merged PDF", http.StatusInternalServerError)
		return
	}
	session.Mutex.Lock()
	session.OutputFile = outputPath
	session.MergeStatus = "done"
	session.Mutex.Unlock()
	downloadURL := fmt.Sprintf("/api/sessions/%s/files/%s", sessionID, outputFilename)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"downloadUrl": "%s"}`, downloadURL)
}

// DownloadFile godoc
// @Summary      Download merged PDF
// @Description  Downloads the merged PDF file for the session
// @Tags         files
// @Produce      application/pdf
// @Param        sessionID  path      string  true  "Session ID"
// @Param        filename   path      string  true  "Merged PDF filename"
// @Success      200  {file}  file  "PDF file download"
// @Failure      403  {string}  string  "Unauthorized access to file"
// @Failure      404  {string}  string  "Session or file not found"
// @Router       /api/sessions/{sessionID}/files/{filename} [get]
func (h *APIHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	filename := chi.URLParam(r, "filename")
	session, exists := h.SessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	filepath := filepath.Join(h.OutputDir, filename)
	if session.OutputFile != filepath {
		http.Error(w, "Unauthorized access to file", http.StatusForbidden)
		return
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\"merged.pdf\"")
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, filepath)
	go func() {
		time.Sleep(1 * time.Second)
		session.Cleanup()
		h.SessionManager.DeleteSession(sessionID)
	}()
}

// SignPDF godoc
// @Summary      Sign a PDF file
// @Description  Places a previously uploaded signature image on a PDF at the exact coordinates
// @Tags         signature
// @Accept       json
// @Produce      json
// @Param        sessionID  path    string  true   "Session ID"
// @Param        request    body    object  true   "Sign request"
// @Success      200  {object}  map[string]string  "{ downloadUrl: string }"
// @Failure      400  {string}  string  "Bad request"
// @Failure      404  {string}  string  "Session not found"
// @Router       /api/sessions/{sessionID}/sign [post]
func (h *APIHandler) SignPDF(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	session, exists := h.SessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Parse JSON request
	var req struct {
		SourcePDF string  `json:"sourcePdf"` // Filename only
		Signature string  `json:"signature"` // Filename only
		Page      int     `json:"page"`
		X         float64 `json:"x"`
		Y         float64 `json:"y"`
		Scale     float64 `json:"scale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Signature == "" || req.Page < 1 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	if req.Scale == 0 {
		req.Scale = 1.0
	}

	// Get source PDF path
	var sourcePDFPath string
	if req.SourcePDF == "" {
		http.Error(w, "PDF not specified", http.StatusBadRequest)
		return
	} else {
		sourcePDFPath = filepath.Join(h.UploadDir, req.SourcePDF)

		// Verify the file exists and belongs to this session
		pdfExists := slices.Contains(session.GetFiles(), sourcePDFPath)
		if !pdfExists {
			http.Error(w, "Source PDF not found in session", http.StatusNotFound)
			return
		}
	}

	// Verify signature file exists
	sigPath := filepath.Join(h.UploadDir, req.Signature)

	sigExists := slices.Contains(session.GetFiles(), sigPath)
	if !sigExists {
		http.Error(w, "Signature file not found in session", http.StatusNotFound)
		return
	}

	// Create output file
	signedFilename := fmt.Sprintf("signed-%s.pdf", utils.GenerateUUID())
	signedPath := filepath.Join(h.OutputDir, signedFilename)

	// Apply signature
	if err := pdf.SignPDF(sourcePDFPath, sigPath, req.Page, req.X, req.Y, req.Scale, signedPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply signature: %v", err), http.StatusInternalServerError)
		return
	}

	// Update session with new output file
	session.Mutex.Lock()
	if session.OutputFile != "" {
		log.Printf("Removing old output file: %s", session.OutputFile)
		os.Remove(session.OutputFile)
	}
	session.OutputFile = signedPath
	session.Mutex.Unlock()

	// Return download URL
	downloadURL := fmt.Sprintf("/api/sessions/%s/files/%s", sessionID, signedFilename)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"downloadUrl": "%s"}`, downloadURL)
}

// UploadSignature godoc
// @Summary      Upload a signature image
// @Description  Uploads a signature image (PNG/JPEG) to the session
// @Tags         signature
// @Accept       multipart/form-data
// @Produce      json
// @Param        sessionID  path      string  true  "Session ID"
// @Param        signature  formData  file    true  "Signature image file (PNG/JPEG)"
// @Success      200  {object}  map[string]interface{}  "{ filename: string, size: int }"
// @Failure      400  {string}  string  "Bad request - invalid image format"
// @Failure      404  {string}  string  "Session not found"
// @Router       /api/sessions/{sessionID}/signature [post]
func (h *APIHandler) UploadSignature(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	session, exists := h.SessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	const maxUploadSize = 5 * 1024 * 1024 // 5MB max for signature images
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("signature")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		http.Error(w, "Only PNG and JPEG images are allowed", http.StatusBadRequest)
		return
	}

	// Read first few bytes to verify it's an image
	header := make([]byte, 512)
	if _, err := file.Read(header); err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(header)

	// Check that the content type is specifically PNG or JPEG/JPG
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	if !allowedTypes[contentType] {
		http.Error(w, "Invalid image format. Only PNG and JPEG images are allowed", http.StatusBadRequest)
		return
	}

	// Additional security check: verify extension matches detected content type
	extForType := strings.ToLower(filepath.Ext(handler.Filename))
	validExtensions := map[string][]string{
		"image/jpeg": {".jpg", ".jpeg"},
		"image/png":  {".png"},
	}

	isValidExt := false
	if extensions, ok := validExtensions[contentType]; ok {
		isValidExt = slices.Contains(extensions, extForType)
	}

	if !isValidExt {
		http.Error(w, "File extension doesn't match content type", http.StatusBadRequest)
		return
	}

	sanitizedFilename := utils.SanitizeFilename(handler.Filename)
	filename := fmt.Sprintf("sig-%s-%s", utils.GenerateUUID(), sanitizedFilename)
	filepath := filepath.Join(h.UploadDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Add signature file reference to session
	session.AddFile(filepath)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"filename": "%s", "size": %d}`, filename, handler.Size)
}
