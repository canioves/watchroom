package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/watchroom/backend/internal/room"
)

type RoomHandler struct {
	manager *room.Manager
}

func NewRoomHandler(manager *room.Manager) *RoomHandler {
	return &RoomHandler{manager: manager}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	rm := h.manager.Create()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"roomId": rm.ID})
}

func (h *RoomHandler) Check(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["roomID"]
	_, ok := h.manager.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"roomId": id})
}
