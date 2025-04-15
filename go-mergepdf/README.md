# go-mergepdf

A simple Go web service for merging PDF files via a REST API.

## Features
- Upload multiple PDF files in a session
- Reorder uploaded files before merging
- Merge PDFs into a single file
- Download the merged PDF
- Automatic cleanup of uploaded and merged files

## API Endpoints

All endpoints are prefixed with `/api/sessions`.

### 1. Create a Session
- **POST** `/api/sessions/`
- **Response:**
  ```json
  { "sessionId": "<session-id>" }
  ```

### 2. Upload a PDF File
- **POST** `/api/sessions/{sessionID}/files`
- **Body:** `multipart/form-data` with a `pdf` file field
- **Response:**
  ```json
  { "filename": "upload/<stored-filename>", "size": 12345 }
  ```

### 3. Set File Order
- **PUT** `/api/sessions/{sessionID}/order`
- **Body:**
  ```json
  { "files": ["upload/<filename1>", "upload/<filename2>", ...] }
  ```
- **Response:**
  ```json
  { "success": true }
  ```

### 4. Merge Files
- **POST** `/api/sessions/{sessionID}/actions/merge`
- **Response:**
  ```json
  { "downloadUrl": "/api/sessions/{sessionID}/files/merged-<uuid>.pdf" }
  ```

### 5. Download Merged PDF
- **GET** `/api/sessions/{sessionID}/files/{filename}`
- **Response:**
  - Content-Type: `application/pdf`
  - Content-Disposition: `attachment; filename="merged.pdf"`

## Project Structure
- `cmd/api/main.go` - Application entrypoint
- `internal/handlers/` - HTTP handlers
- `internal/session/` - Session management
- `internal/pdf/` - PDF operations
- `internal/server/` - Server and routing
- `internal/utils/` - Utility functions
