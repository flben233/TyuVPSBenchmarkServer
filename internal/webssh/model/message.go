package model

type MessageType string

const (
	TypeConnect               MessageType = "connect"
	TypeInput                 MessageType = "input"
	TypeResize                MessageType = "resize"
	TypePing                  MessageType = "ping"
	TypeAgentTask             MessageType = "agent_task"
	TypeAgentMessage          MessageType = "agent_message"
	TypeAgentApprovalResponse MessageType = "agent_approval_response"
	TypeOutput                MessageType = "output"
	TypeConnected             MessageType = "connected"
	TypeError                 MessageType = "error"
	TypeClosed                MessageType = "closed"
	TypeAgentMessageStart     MessageType = "agent_message_start"
	TypeAgentToken            MessageType = "agent_token"
	TypeAgentMessageEnd       MessageType = "agent_message_end"
	TypeAgentState            MessageType = "agent_state"

	// Backward-compatible aliases.
	TypeAgentMsg MessageType = TypeAgentMessage
	TypeAgentAck MessageType = TypeAgentApprovalResponse
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
	Type         MessageType `json:"type"`
	Data         string      `json:"data,omitempty"`
	Message      string      `json:"message,omitempty"`
	TaskID       string      `json:"task_id,omitempty"`
	MessageID    string      `json:"message_id,omitempty"`
	Delta        string      `json:"delta,omitempty"`
	FinishReason string      `json:"finish_reason,omitempty"`
	State        string      `json:"state,omitempty"`
}
