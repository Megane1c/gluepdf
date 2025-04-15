// Package session manages user sessions and file lists for PDF merging.
//
// Types:
//   - Session: Tracks uploaded files, output file, and status for a user session.
//   - SessionManager: Manages all active sessions.
//
// Expected outputs:
// - Session IDs are unique (UUID)
// - Files are tracked per session
// - Cleanup removes all files for a session
//
// Used by API handlers to manage user state.
package session

import (
	"go-mergepdf/internal/utils"
	"os"
	"sync"
	"time"
)

type Session struct {
	ID          string
	Files       []string
	OutputFile  string
	CreatedAt   time.Time
	MergeStatus string
	Mutex       sync.Mutex
}

type SessionManager struct {
	Sessions map[string]*Session
	Mutex    sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		Sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession() *Session {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	session := &Session{
		ID:          utils.GenerateUUID(),
		Files:       []string{},
		CreatedAt:   time.Now(),
		MergeStatus: "idle",
	}
	sm.Sessions[session.ID] = session
	return session
}

func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()
	session, exists := sm.Sessions[id]
	return session, exists
}

func (sm *SessionManager) DeleteSession(id string) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()
	delete(sm.Sessions, id)
}

func (s *Session) AddFile(filepath string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Files = append(s.Files, filepath)
}

func (s *Session) SetFiles(files []string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Files = files
}

func (s *Session) GetFiles() []string {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	return s.Files
}

func (s *Session) Cleanup() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for _, file := range s.Files {
		os.Remove(file)
	}
	if s.OutputFile != "" {
		os.Remove(s.OutputFile)
	}
}
