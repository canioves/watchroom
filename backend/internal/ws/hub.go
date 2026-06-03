package ws

import (
	"log"
	"sync"
	"time"

	"github.com/watchroom/backend/internal/room"
)

type Hub struct {
	clients map[string]map[*room.Client]struct{} // roomID → clients
	timers  map[string]*time.Timer
	manager *room.Manager
	mu      sync.RWMutex
}

func NewHub(manager *room.Manager) *Hub {
	return &Hub{
		clients: make(map[string]map[*room.Client]struct{}),
		timers:  make(map[string]*time.Timer),
		manager: manager,
	}
}

func (h *Hub) Register(c *room.Client) {
	h.mu.Lock()
	if t, ok := h.timers[c.RoomID]; ok {
		t.Stop()
		delete(h.timers, c.RoomID)
	}
	if h.clients[c.RoomID] == nil {
		h.clients[c.RoomID] = make(map[*room.Client]struct{})
	}
	h.clients[c.RoomID][c] = struct{}{}
	h.mu.Unlock()
	log.Printf("[join] user=%s room=%s", c.Nickname, c.RoomID)
}

func (h *Hub) Unregister(c *room.Client) {
	h.mu.Lock()
	delete(h.clients[c.RoomID], c)
	if len(h.clients[c.RoomID]) == 0 {
		t := time.AfterFunc(10*time.Second, func() {
			h.mu.Lock()
			delete(h.timers, c.RoomID)
			h.mu.Unlock()
			h.manager.Delete(c.RoomID)
		})
		h.timers[c.RoomID] = t
	}
	h.mu.Unlock()
	close(c.Send)
	log.Printf("[leave] user=%s room=%s", c.Nickname, c.RoomID)
}

// Broadcast sends msg to all clients in the room except sender (nil = send to all).
func (h *Hub) Broadcast(roomID string, msg []byte, except *room.Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[roomID] {
		if c == except {
			continue
		}
		select {
		case c.Send <- msg:
		default:
			log.Printf("[hub] slow client %s, dropping message", c.Nickname)
		}
	}
}

func (h *Hub) Participants(roomID string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.clients[roomID]
	result := make([]Message, 0, len(clients))
	for c := range clients {
		result = append(result, Message{
			Type:     UserJoin,
			UserID:   c.ID,
			Nickname: c.Nickname,
		})
	}
	return result
}
