package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/watchroom/backend/internal/room"
	"github.com/watchroom/backend/internal/ws"
)

func NewRouter(manager *room.Manager, hub *ws.Hub, allowedOrigin string) http.Handler {
	r := mux.NewRouter()

	roomHandler := NewRoomHandler(manager)
	wsHandler := ws.NewHandler(hub, manager, allowedOrigin)

	r.HandleFunc("/api/rooms", roomHandler.Create).Methods(http.MethodPost)
	r.HandleFunc("/api/rooms/{roomID}", roomHandler.Check).Methods(http.MethodGet)
	r.HandleFunc("/ws/{roomID}", wsHandler.ServeWS)

	r.Use(corsMiddleware(allowedOrigin))

	return r
}

func corsMiddleware(allowedOrigin string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
