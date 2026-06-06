package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/watchroom/backend/internal/room"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 4096
)

type Handler struct {
	hub      *Hub
	manager  *room.Manager
	upgrader websocket.Upgrader
}

func NewHandler(hub *Hub, manager *room.Manager, allowedOrigin string) *Handler {
	return &Handler{
		hub:     hub,
		manager: manager,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == allowedOrigin || origin == ""
			},
		},
	}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomID := mux.Vars(r)["roomID"]
	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		nickname = "anon"
	}
	rm, ok := h.manager.Get(roomID)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}

	clientID := uuid.NewString()[:8]
	client := room.NewClient(clientID, nickname, roomID, conn)
	existing := h.hub.Participants(roomID)
	h.hub.Register(client)

	// send current state to the new participant
	state := rm.GetState()
	if state.IsPlaying {
		elapsed := float64(time.Now().UnixMilli()-state.UpdatedAt) / 1000
		state.Position += elapsed
	}
	syncMsg, _ := json.Marshal(Message{
		Type:      StateSync,
		VideoID:   state.VideoID,
		Position:  state.Position,
		IsPlaying: state.IsPlaying,
		Nickname:  nickname,
		UserID:    clientID,
	})

	client.Send <- syncMsg

	for _, p := range existing {
		msg, _ := json.Marshal(p)
		client.Send <- msg
	}

	// broadcast join to others
	joinMsg, _ := json.Marshal(Message{Type: UserJoin, UserID: clientID, Nickname: nickname})
	h.hub.Broadcast(roomID, joinMsg, client)

	go h.writePump(client)
	h.readPump(client, rm)
}

func (h *Handler) readPump(c *room.Client, rm *room.Room) {
	defer func() {
		leaveMsg, _ := json.Marshal(Message{Type: UserLeave, UserID: c.ID, Nickname: c.Nickname})
		h.hub.Broadcast(c.RoomID, leaveMsg, c)
		h.hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMsgSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ws] read error user=%s: %v", c.Nickname, err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ws] bad message from %s: %v", c.Nickname, err)
			continue
		}
		msg.UserID = c.ID
		switch msg.Type {
		case PlayerPlay:
			rm.SetState(room.State{
				VideoID:   rm.GetState().VideoID,
				Position:  msg.Position,
				IsPlaying: true,
				UpdatedAt: time.Now().UnixMilli(),
			})
		case PlayerPause:
			rm.SetState(room.State{
				VideoID:   rm.GetState().VideoID,
				Position:  msg.Position,
				IsPlaying: false,
				UpdatedAt: time.Now().UnixMilli(),
			})
		case VideoLoad:
			rm.SetState(room.State{
				VideoID:   msg.VideoID,
				Position:  0,
				IsPlaying: true,
				UpdatedAt: time.Now().UnixMilli(),
			})
		case MessageChat:
			msg.Nickname = c.Nickname
		}

		log.Printf("[msg] type=%s user=%s room=%s", msg.Type, c.Nickname, c.RoomID)

		out, _ := json.Marshal(msg)
		h.hub.Broadcast(c.RoomID, out, c)
	}
}

func (h *Handler) writePump(c *room.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
