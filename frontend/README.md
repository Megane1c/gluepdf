# Frontend

This is the frontend (React + TypeScript + Vite) for a PDF merging application. It interacts with a backend API to handle session creation, PDF uploads, reordering, merging, and downloading merged files.

## Features
- Create a new session
- Upload multiple PDF files
- Reorder uploaded files
- Merge the files into a single PDF
- Download the merged file

## API Integration

This project uses `axios` to communicate with a backend running at `http://localhost:8080`. Below are the main API calls handled by the app:

### `createSession()`

**POST** `/api/sessions/`

Creates a new user session by sending relevant authentication data to the backend.

### `uploadFile(sessionId, file)`
**POST** `/api/sessions/:sessionId/files`

Uploads a PDF file to the specified session.

### `updateOrder(sessionId, files)`
**PUT** `/api/sessions/:sessionId/order`

Updates the order of uploaded files.

### `mergeFiles(sessionId)`
**POST** `/api/sessions/:sessionId/actions/merge`

Triggers the merge action and returns a download URL.

### `downloadFile(url)`
Downloads the merged file and refreshes the page.

## Project Structure
- `public/` - Static assets
- `src/` - Main application source code
- `src/components/` - React components such as `PDFList` and `PDFMerger`
- `src/services/` - Backend communication via `api.ts`
- `App.tsx` - Root component
- `main.tsx` - React + Vite entry point
- `App.css`, `index.css` - Component and global styles
- `index.html` - HTML template used by Vite
- `README.md` - Project documentation (this file)