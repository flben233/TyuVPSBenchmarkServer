package service

import (
	"log"
	"sync"

	"VPSBenchmarkBackend/internal/webssh/model"
)

type AgentStreamSender func(model.ServerMessage)

type agentStreamSlot struct {
	mu      sync.Mutex
	cond    *sync.Cond
	ownerID string
	sender  AgentStreamSender
	version uint64
	active  int
}

type AgentStreamBridge struct {
	mu      sync.RWMutex
	senders map[string]*agentStreamSlot
}

var agentStreamBridgePanicHook func(taskID string, recovered any)

func NewAgentStreamBridge() *AgentStreamBridge {
	return &AgentStreamBridge{
		senders: make(map[string]*agentStreamSlot),
	}
}

func newAgentStreamSlot(ownerID string, sender AgentStreamSender) *agentStreamSlot {
	slot := &agentStreamSlot{ownerID: ownerID, sender: sender}
	slot.cond = sync.NewCond(&slot.mu)
	return slot
}

func (b *AgentStreamBridge) Register(taskID string, ownerID string, sender AgentStreamSender) {
	if b == nil || taskID == "" || ownerID == "" || sender == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if slot, ok := b.senders[taskID]; ok && slot != nil {
		slot.mu.Lock()
		slot.version++
		slot.ownerID = ownerID
		slot.sender = sender
		slot.mu.Unlock()
		return
	}
	b.senders[taskID] = newAgentStreamSlot(ownerID, sender)
}

func (b *AgentStreamBridge) Unregister(taskID string, ownerID string) {
	if b == nil || taskID == "" || ownerID == "" {
		return
	}

	b.mu.Lock()
	slot, ok := b.senders[taskID]
	if ok && slot != nil {
		slot.mu.Lock()
		if slot.ownerID != ownerID {
			slot.mu.Unlock()
			b.mu.Unlock()
			return
		}
		slot.mu.Unlock()
		delete(b.senders, taskID)
	}
	b.mu.Unlock()
	if !ok || slot == nil {
		return
	}

	slot.mu.Lock()
	slot.version++
	slot.sender = nil
	for slot.active > 0 {
		slot.cond.Wait()
	}
	slot.mu.Unlock()
}

func (b *AgentStreamBridge) EmitMessageStart(taskID string, messageID string) {
	b.emit(taskID, model.ServerMessage{
		Type:      model.TypeAgentMessageStart,
		TaskID:    taskID,
		MessageID: messageID,
	})
}

func (b *AgentStreamBridge) EmitToken(taskID string, messageID string, delta string) {
	b.emit(taskID, model.ServerMessage{
		Type:      model.TypeAgentToken,
		TaskID:    taskID,
		MessageID: messageID,
		Delta:     delta,
	})
}

func (b *AgentStreamBridge) EmitMessageEnd(taskID string, messageID string, reason string) {
	b.emit(taskID, model.ServerMessage{
		Type:         model.TypeAgentMessageEnd,
		TaskID:       taskID,
		MessageID:    messageID,
		FinishReason: reason,
	})
}

func (b *AgentStreamBridge) EmitState(taskID string, state string, message string) {
	b.emit(taskID, model.ServerMessage{
		Type:    model.TypeAgentState,
		TaskID:  taskID,
		State:   state,
		Message: message,
	})
}

func (b *AgentStreamBridge) emit(taskID string, message model.ServerMessage) {
	if b == nil || taskID == "" {
		return
	}

	b.mu.RLock()
	slot, ok := b.senders[taskID]
	b.mu.RUnlock()
	if !ok || slot == nil {
		return
	}

	slot.mu.Lock()
	version := slot.version
	sender := slot.sender
	if sender == nil {
		slot.mu.Unlock()
		return
	}
	slot.active++
	slot.mu.Unlock()

	defer func() {
		slot.mu.Lock()
		slot.active--
		if slot.active == 0 {
			slot.cond.Broadcast()
		}
		slot.mu.Unlock()
	}()
	defer func() {
		if recovered := recover(); recovered != nil {
			reportAgentStreamBridgePanic(taskID, version, recovered)
		}
	}()

	sender(message)
}

func reportAgentStreamBridgePanic(taskID string, version uint64, recovered any) {
	if agentStreamBridgePanicHook != nil {
		agentStreamBridgePanicHook(taskID, recovered)
		return
	}
	log.Printf("agent stream bridge sender panic: task_id=%s version=%d recovered=%v", taskID, version, recovered)
}
