package service

import (
	"VPSBenchmarkBackend/internal/exporter"
	"VPSBenchmarkBackend/internal/mq"
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TracertRequest struct {
	Target string `json:"target" binding:"required"`
	Mode   string `json:"mode" binding:"required,oneof=icmp tcp"`
	Port   uint16 `json:"port"`
}

type TracertResponse struct {
	TaskID string `json:"task_id"`
}

const tracertSource = "tool_tracert"

func init() {
	mq.LateSubscribe(tracertSource, context.Background(), postTracert)
}

func postTracert(msg *amqp.Delivery) error {
	var resp exporter.TracertResp
	if err := json.Unmarshal(msg.Body, &resp); err != nil {
		return err
	}
	err := mq.SetTask(mq.Task[map[string]interface{}]{
		ID:       msg.MessageId,
		Status:   mq.TaskDone,
		Progress: 1.0,
		Result:   resp.Result,
	})
	if err != nil {
		return err
	}
	return nil
}

func Traceroute(req *TracertRequest) (string, error) {
	taskID, err := mq.PublishJSONWithID(exporter.TracertRoute, tracertSource, exporter.TracertReq{
		Mode:   req.Mode,
		Target: req.Target,
		Port:   uint64(req.Port),
	}, "")
	if err != nil {
		return "", err
	}
	err = mq.SetTask(mq.Task[any]{
		ID:       taskID,
		Status:   mq.TaskPending,
		Progress: 0.0,
		Result:   nil,
	})
	if err != nil {
		return "", err
	}
	return taskID, nil
}
