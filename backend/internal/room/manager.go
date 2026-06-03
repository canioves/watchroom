package room

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

func (m *Manager) Create() *Room {
	id := uuid.NewString()[:8]
	r := New(id)
	m.mu.Lock()
	m.rooms[id] = r
	m.mu.Unlock()
	log.Printf("[room] %s created", id)
	return r
}

func (m *Manager) Get(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[id]
	return r, ok
}

func (m *Manager) Delete(id string) {
	m.mu.Lock()
	delete(m.rooms, id)
	m.mu.Unlock()
	log.Printf("[room] %s deleted", id)
}
