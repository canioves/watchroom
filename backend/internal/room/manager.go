package room

import (
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
	return r
}

func (m *Manager) Get(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[id]
	return r, ok
}
