# PDF Merge App

A web application for merging PDF files. This monorepo contains the backend (Go) and will include a frontend (e.g., React, Vue) in the future.

## Project Structure

```
go-mergepdf/      # Go backend (REST API for PDF merging)
frontend/         # (Planned) Frontend web application
```

- **go-mergepdf/**: Contains the Go backend service for uploading, merging, and downloading PDF files.  
- **frontend/**: (To be added) Will contain the frontend client for interacting with the backend.

## Getting Started

### Backend (Go)

1. **Prerequisites:**  
   - Go 1.20 or newer

2. **Setup:**  
   ```bash
   cd go-mergepdf
   go mod tidy
   go run cmd/api/main.go
   ```

The server will start on port 8080 by default.

For API details, see [go-mergepdf/README.md](go-mergepdf/README.md)

### API Documentation

Interactive API documentation is available via Swagger:

- [Swagger UI](http://localhost:8080/swagger/index.html)

### Frontend (Planned)

The frontend will be added in the `frontend/` directory. Setup instructions will be provided once available.

## Acknowledgements
- Backend project structure generated with [go-blueprint](https://github.com/Melkeydev/go-blueprint)