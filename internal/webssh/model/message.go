package model

type MessageType string

const (
	TypeConnect       MessageType = "connect"
	TypeInput         MessageType = "input"
	TypeResize        MessageType = "resize"
	TypePing          MessageType = "ping"
	TypeAgentTask     MessageType = "agent_task"
	TypeAgentMsg      MessageType = "agent_message"
	TypeAgentAck      MessageType = "agent_approval_response"
	TypeOutput        MessageType = "output"
	TypeConnected     MessageType = "connected"
	TypeError         MessageType = "error"
	TypeClosed        MessageType = "closed"
	TypeAgentUpdate   MessageType = "agent_update"
	TypeAgentApproval MessageType = "agent_approval"
	TypeAgentError    MessageType = "agent_error"
	TypeAgentDone     MessageType = "agent_done"
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
	TaskID     string      `json:"task_id,omitempty"`
	Message    string      `json:"message,omitempty"`
	Approved   *bool       `json:"approved,omitempty"`
}

type ServerMessage struct {
	Type     MessageType `json:"type"`
	Data     string      `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	TaskID   string      `json:"task_id,omitempty"`
	Status   string      `json:"status,omitempty"`
	Question string      `json:"question,omitempty"`
	Summary  string      `json:"summary,omitempty"`
}
