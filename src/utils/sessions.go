package utils

import (
	"crypto/rand"
	"encoding/hex"
	"potato-bones/src/globals"
	"sync"
	"time"
)

type Session struct {
	ID string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mutex sync.Mutex
	duration time.Duration
}

func generateID() string {
	data := make([]byte, 16)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(data)
}

func (manager *SessionManager) CreateSession() *Session {
	manager.mutex.Lock(); defer manager.mutex.Unlock()

	id := generateID()
	tick := time.Now()
	session := &Session{
		ID:        id,
		CreatedAt: tick,
		ExpiresAt: tick.Add(manager.duration),
	}
	manager.sessions[id] = session
	return session
}

func (manager *SessionManager) GetSession(id string) (*Session, bool) {
	manager.mutex.Lock(); defer manager.mutex.Unlock()

	session, ok := manager.sessions[id]
	if !ok || time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

func (manager *SessionManager) KillSession(id string) {
	manager.mutex.Lock(); defer manager.mutex.Unlock()
	delete(manager.sessions, id)
}

func (manager *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		manager.mutex.Lock()
		now := time.Now()

		for id, session := range manager.sessions {
			if now.After(session.ExpiresAt) {
				delete(manager.sessions, id)
			}
		}

		manager.mutex.Unlock()
	}
}

func NewSessionManager() *SessionManager {
	manager := &SessionManager{
		sessions: make(map[string]*Session),
		duration: time.Minute * time.Duration(*globals.SessionLength),
	}

	go manager.cleanupLoop()
	return manager
}