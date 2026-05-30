package ws

import (
	"log"
	"sync"

	"github.com/watchroom/backend/internal/room"
)

type Hub struct {
	clients map[string]map[*room.Client]struct{} // roomID → clients
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*room.Client]struct{})}
}

func (h *Hub) Register(c *room.Client) {
	h.mu.Lock()
	if h.clients[c.RoomID] == nil {
		h.clients[c.RoomID] = make(map[*room.Client]struct{})
	}
	h.clients[c.RoomID][c] = struct{}{}
	h.mu.Unlock()
	log.Printf("[join] user=%s room=%s", c.ID, c.RoomID)
}

func (h *Hub) Unregister(c *room.Client) {
	h.mu.Lock()
	delete(h.clients[c.RoomID], c)
	h.mu.Unlock()
	close(c.Send)
	log.Printf("[leave] user=%s room=%s", c.ID, c.RoomID)
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
			log.Printf("[hub] slow client %s, dropping message", c.ID)
		}
	}
}
