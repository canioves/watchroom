package room

import "sync"

type State struct {
	VideoID   string  `json:"videoId"`
	Position  float64 `json:"position"`
	IsPlaying bool    `json:"isPlaying"`
	UpdatedAt int64   `json:"updatedAt"`
}

type Room struct {
	ID    string
	State State
	mu    sync.RWMutex
}

func New(id string) *Room {
	return &Room{ID: id}
}

func (r *Room) GetState() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.State
}

func (r *Room) SetState(s State) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.State = s
}
