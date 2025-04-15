package server

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-mergepdf/internal/session"
)

func setupTestServer() *httptest.Server {
	s := &Server{
		SessionManager: session.NewSessionManager(),
		UploadDir:      "uploads",
		OutputDir:      "output",
	}
	return httptest.NewServer(s.RegisterRoutes())
}

func teardownUploadsAndOutput() {
	dirs := []string{"uploads", "output"}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				_ = os.Remove(dir + "/" + entry.Name())
			}
		}
	}
}

func TestMain(m *testing.M) {
	teardownUploadsAndOutput() // Clean before tests
	os.Chdir("../../")         // Change to project root
	code := m.Run()
	teardownUploadsAndOutput() // Clean after tests
	os.Exit(code)
}

func TestCreateSession(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/sessions/", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if result["sessionId"] == "" {
		t.Error("Expected sessionId in response")
	}
}

func TestUploadFile(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create session first
	resp, _ := http.Post(server.URL+"/api/sessions/", "application/json", nil)
	var result map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&result)
	sessionID := result["sessionId"]

	t.Run("valid PDF", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		file, err := os.Open("testfiles/DICT-ICT014.pdf")
		if err != nil {
			t.Fatalf("Failed to open valid test PDF: %v", err)
		}
		defer file.Close()
		part, _ := writer.CreateFormFile("pdf", filepath.Base(file.Name()))
		_, _ = io.Copy(part, file)
		writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/api/sessions/"+sessionID+"/files", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp2, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to upload file: %v", err)
		}
		defer resp2.Body.Close()
		body, _ := io.ReadAll(resp2.Body)
		t.Logf("Response body: %s", string(body))
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", resp2.StatusCode)
		}
	})

	t.Run("invalid PDF", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		file, err := os.Open("testfiles/notpdf.pdf")
		if err != nil {
			t.Fatalf("Failed to open invalid test file: %v", err)
		}
		defer file.Close()
		part, _ := writer.CreateFormFile("pdf", filepath.Base(file.Name()))
		_, _ = io.Copy(part, file)
		writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/api/sessions/"+sessionID+"/files", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp2, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to upload file: %v", err)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusOK {
			t.Fatalf("Expected error status for invalid PDF, got %d", resp2.StatusCode)
		}
	})
}

func TestMergeFiles(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create session
	resp, _ := http.Post(server.URL+"/api/sessions/", "application/json", nil)
	var result map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&result)
	sessionID := result["sessionId"]

	// Upload two files
	for _, fname := range []string{"DICT-ICT014.pdf", "DICT-WD001.pdf"} {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		file, _ := os.Open("testfiles/" + fname)
		defer file.Close()
		part, _ := writer.CreateFormFile("pdf", fname)
		_, _ = io.Copy(part, file)
		writer.Close()
		req, _ := http.NewRequest("POST", server.URL+"/api/sessions/"+sessionID+"/files", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		http.DefaultClient.Do(req)
	}

	// Merge
	req, _ := http.NewRequest("POST", server.URL+"/api/sessions/"+sessionID+"/actions/merge", nil)
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to merge files: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp3.StatusCode)
	}
	var mergeResult map[string]string
	_ = json.NewDecoder(resp3.Body).Decode(&mergeResult)
	if !strings.Contains(mergeResult["downloadUrl"], "/api/sessions/") {
		t.Error("Expected downloadUrl in response")
	}
}
