package ws

type MessageType string

const (
	PlayerPlay  MessageType = "play"
	PlayerPause MessageType = "pause"
	PlayerSeek  MessageType = "seek"
	VideoLoad   MessageType = "load"
	StateSync   MessageType = "sync"
	UserJoin    MessageType = "join"
	UserLeave   MessageType = "leave"
)

type Message struct {
	Type      MessageType `json:"type"`
	VideoID   string      `json:"videoId,omitempty"`
	Position  float64     `json:"position,omitempty"`
	SentAt    int64       `json:"sentAt,omitempty"`
	UserID    string      `json:"userId,omitempty"`
	IsPlaying bool        `json:"isPlaying,omitempty"`
}
