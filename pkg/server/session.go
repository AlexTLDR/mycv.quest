package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

type Session struct {
	ID        string
	CreatedAt time.Time
	PDFData   map[string][]byte // templateKey -> PDF content
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
	}

	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()

	return sm
}

func (sm *SessionManager) GetOrCreateSession(r *http.Request) *Session {
	// Try to get session from cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		if session := sm.GetSession(cookie.Value); session != nil {
			return session
		}
	}

	// Create new session
	return sm.CreateSession()
}

func (sm *SessionManager) CreateSession() *Session {
	sessionID := generateSessionID()

	session := &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		PDFData:   make(map[string][]byte),
	}

	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.mutex.Unlock()

	return session
}

func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.sessions[sessionID]
}

func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()
}

func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, session *Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1 hour
	})
}

func (sm *SessionManager) StorePDF(sessionID, templateKey string, pdfData []byte) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.PDFData[templateKey] = pdfData
	}
}

func (sm *SessionManager) GetPDF(sessionID, templateKey string) ([]byte, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if session, exists := sm.sessions[sessionID]; exists {
		if pdfData, exists := session.PDFData[templateKey]; exists {
			return pdfData, true
		}
	}
	return nil, false
}

func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			// Remove sessions older than 1 hour
			if now.Sub(session.CreatedAt) > time.Hour {
				delete(sm.sessions, id)
			}
		}
		sm.mutex.Unlock()
	}
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
