package service

import (
	"sync/atomic"
	"testing"
	"time"

	"VPSBenchmarkBackend/internal/webssh/model"
)

func TestStreamBridgeRoutesTokenByTaskID(t *testing.T) {
	bridge := NewAgentStreamBridge()

	var task1Messages []model.ServerMessage
	var task2Messages []model.ServerMessage

	bridge.Register("task-1", "owner-1", func(msg model.ServerMessage) {
		task1Messages = append(task1Messages, msg)
	})
	bridge.Register("task-2", "owner-2", func(msg model.ServerMessage) {
		task2Messages = append(task2Messages, msg)
	})

	bridge.EmitToken("task-1", "message-1", "hello")

	if len(task1Messages) != 1 {
		t.Fatalf("expected 1 message for task-1, got %d", len(task1Messages))
	}
	if len(task2Messages) != 0 {
		t.Fatalf("expected 0 messages for task-2, got %d", len(task2Messages))
	}

	got := task1Messages[0]
	if got.Type != model.TypeAgentToken {
		t.Fatalf("expected type %q, got %q", model.TypeAgentToken, got.Type)
	}
	if got.TaskID != "task-1" {
		t.Fatalf("expected task_id task-1, got %q", got.TaskID)
	}
	if got.MessageID != "message-1" {
		t.Fatalf("expected message_id message-1, got %q", got.MessageID)
	}
	if got.Delta != "hello" {
		t.Fatalf("expected delta hello, got %q", got.Delta)
	}
}

func TestStreamBridgeUnregisterAndUnknownTaskNoOp(t *testing.T) {
	bridge := NewAgentStreamBridge()

	var messages []model.ServerMessage
	bridge.Register("task-1", "owner-1", func(msg model.ServerMessage) {
		messages = append(messages, msg)
	})

	bridge.Unregister("task-1", "owner-1")
	bridge.EmitToken("task-1", "message-1", "should-not-deliver")
	bridge.EmitState("unknown-task", "thinking", "planning")

	if len(messages) != 0 {
		t.Fatalf("expected no messages after unregister and unknown task emits, got %d", len(messages))
	}
}

func TestStreamBridgeUnregisterWaitsForInFlightEmitAndStopsSubsequentDelivery(t *testing.T) {
	bridge := NewAgentStreamBridge()

	started := make(chan struct{})
	release := make(chan struct{})
	emitDone := make(chan struct{})

	var delivered int32
	bridge.Register("task-1", "owner-1", func(msg model.ServerMessage) {
		atomic.AddInt32(&delivered, 1)
		if msg.Delta == "first" {
			close(started)
			<-release
		}
	})

	go func() {
		bridge.EmitToken("task-1", "message-1", "first")
		close(emitDone)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for in-flight emit to start")
	}

	unregisterDone := make(chan struct{})
	go func() {
		bridge.Unregister("task-1", "owner-1")
		close(unregisterDone)
	}()

	select {
	case <-unregisterDone:
		t.Fatal("expected unregister to wait for in-flight emit")
	case <-time.After(50 * time.Millisecond):
	}

	bridge.EmitToken("task-1", "message-2", "second")
	if got := atomic.LoadInt32(&delivered); got != 1 {
		t.Fatalf("expected no delivery while unregister is pending, got %d", got)
	}

	close(release)

	select {
	case <-emitDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for in-flight emit to finish")
	}

	select {
	case <-unregisterDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for unregister to finish")
	}

	bridge.EmitToken("task-1", "message-3", "third")
	if got := atomic.LoadInt32(&delivered); got != 1 {
		t.Fatalf("expected no delivery after unregister, got %d", got)
	}
}

func TestStreamBridgeSenderPanicIsolated(t *testing.T) {
	bridge := NewAgentStreamBridge()
	originalHook := agentStreamBridgePanicHook
	t.Cleanup(func() {
		agentStreamBridgePanicHook = originalHook
	})

	var hookCalled int32
	var hookTaskID string
	agentStreamBridgePanicHook = func(taskID string, recovered any) {
		if recovered == nil {
			t.Fatal("expected recovered panic value")
		}
		hookTaskID = taskID
		atomic.StoreInt32(&hookCalled, 1)
	}

	bridge.Register("task-1", "owner-1", func(model.ServerMessage) {
		panic("sender panic")
	})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected emit to recover sender panic, got panic: %v", r)
		}
	}()

	bridge.EmitToken("task-1", "message-1", "hello")
	if atomic.LoadInt32(&hookCalled) != 1 {
		t.Fatal("expected panic hook to be called")
	}
	if hookTaskID != "task-1" {
		t.Fatalf("expected panic hook task_id task-1, got %q", hookTaskID)
	}
}

func TestStreamBridgeReentrantUnregisterNoDeadlock(t *testing.T) {
	bridge := NewAgentStreamBridge()

	var delivered int32
	unregisterDone := make(chan struct{})
	emitDone := make(chan struct{})
	bridge.Register("task-1", "owner-1", func(model.ServerMessage) {
		atomic.AddInt32(&delivered, 1)
		go func() {
			bridge.Unregister("task-1", "owner-1")
			close(unregisterDone)
		}()
	})

	go func() {
		bridge.EmitToken("task-1", "message-1", "first")
		close(emitDone)
	}()

	select {
	case <-emitDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for emit to return from reentrant unregister")
	}

	select {
	case <-unregisterDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for reentrant unregister to return")
	}

	bridge.EmitToken("task-1", "message-2", "second")
	if got := atomic.LoadInt32(&delivered); got != 1 {
		t.Fatalf("expected exactly one delivery, got %d", got)
	}
}

func TestStreamBridgeReentrantRegisterNoDeadlock(t *testing.T) {
	bridge := NewAgentStreamBridge()

	var oldDelivered int32
	var newDelivered int32
	emitDone := make(chan struct{})

	newSender := func(model.ServerMessage) {
		atomic.AddInt32(&newDelivered, 1)
	}

	bridge.Register("task-1", "owner-1", func(model.ServerMessage) {
		atomic.AddInt32(&oldDelivered, 1)
		bridge.Register("task-1", "owner-2", newSender)
	})

	go func() {
		bridge.EmitToken("task-1", "message-1", "first")
		close(emitDone)
	}()

	select {
	case <-emitDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for emit to return from reentrant register")
	}

	bridge.EmitToken("task-1", "message-2", "second")
	if got := atomic.LoadInt32(&oldDelivered); got != 1 {
		t.Fatalf("expected old sender to run once, got %d", got)
	}
	if got := atomic.LoadInt32(&newDelivered); got != 1 {
		t.Fatalf("expected new sender to run once, got %d", got)
	}
}

func TestStreamBridgeOwnerScopedUnregisterDoesNotRemoveAnotherConnection(t *testing.T) {
	bridge := NewAgentStreamBridge()

	var owner1Delivered int32
	var owner2Delivered int32

	bridge.Register("task-1", "owner-1", func(model.ServerMessage) {
		atomic.AddInt32(&owner1Delivered, 1)
	})
	bridge.Register("task-1", "owner-2", func(model.ServerMessage) {
		atomic.AddInt32(&owner2Delivered, 1)
	})

	bridge.EmitToken("task-1", "message-1", "first")
	if got := atomic.LoadInt32(&owner1Delivered); got != 0 {
		t.Fatalf("expected owner-1 to be replaced before emit, got %d", got)
	}
	if got := atomic.LoadInt32(&owner2Delivered); got != 1 {
		t.Fatalf("expected owner-2 delivery count 1, got %d", got)
	}

	bridge.Unregister("task-1", "owner-1")
	bridge.EmitToken("task-1", "message-2", "second")
	if got := atomic.LoadInt32(&owner2Delivered); got != 2 {
		t.Fatalf("expected owner-2 sender to remain after stale-owner unregister, got %d", got)
	}

	bridge.Unregister("task-1", "owner-2")
	bridge.EmitToken("task-1", "message-3", "third")
	if got := atomic.LoadInt32(&owner2Delivered); got != 2 {
		t.Fatalf("expected no delivery after active owner unregister, got %d", got)
	}
}
