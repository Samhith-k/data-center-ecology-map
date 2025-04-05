package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

var (
	sessions = make(map[string]string) // sessionID -> username
	mu       sync.RWMutex
)

// GenerateSessionID creates a random 16-byte hex string
func GenerateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "fallback-session-id"
	}
	return hex.EncodeToString(b)
}

// SetUserForSession sets the mapping sessionID -> username
func SetUserForSession(sessionID, username string) {
	mu.Lock()
	defer mu.Unlock()
	sessions[sessionID] = username
}

// GetUserForSession retrieves the username for a given sessionID
func GetUserForSession(sessionID string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	username, ok := sessions[sessionID]
	return username, ok
}

// ClearSession removes a session from the map
func ClearSession(sessionID string) {
	mu.Lock()
	defer mu.Unlock()
	delete(sessions, sessionID)
}
