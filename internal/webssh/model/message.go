package model

type MessageType string

const (
	TypeConnect   MessageType = "connect"
	TypeInput     MessageType = "input"
	TypeResize    MessageType = "resize"
	TypePing      MessageType = "ping"
	TypeOutput    MessageType = "output"
	TypeConnected MessageType = "connected"
	TypeError     MessageType = "error"
	TypeClosed    MessageType = "closed"
)

type ClientMessage struct {
	Type       MessageType `json:"type"`
	Host       string      `json:"host,omitempty"`
	Port       int         `json:"port,omitempty"`
	Username   string      `json:"username,omitempty"`
	Password   string      `json:"password,omitempty"`
	PrivateKey string      `json:"private_key,omitempty"`
	Data       string      `json:"data,omitempty"`
	Cols       int         `json:"cols,omitempty"`
	Rows       int         `json:"rows,omitempty"`
}

type ServerMessage struct {
	Type    MessageType `json:"type"`
	Data    string      `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
