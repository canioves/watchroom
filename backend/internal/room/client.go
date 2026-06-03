package room

import "github.com/gorilla/websocket"

type Client struct {
	ID       string
	Nickname string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan []byte
}

func NewClient(id, nickname, roomID string, conn *websocket.Conn) *Client {
	return &Client{
		ID:       id,
		Nickname: nickname,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 64),
	}
}
