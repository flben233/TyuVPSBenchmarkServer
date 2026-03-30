package service

import (
	"VPSBenchmarkBackend/internal/exporter"
	"VPSBenchmarkBackend/internal/mq"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

type TracertRequest struct {
	Target string `json:"target" binding:"required"`
	Mode   string `json:"mode" binding:"required,oneof=icmp tcp"`
	Port   uint16 `json:"port"`
}

type TracertResponse struct {
	TaskID string `json:"task_id"`
}

var tracertWriter *kafka.Writer

func init() {
	writer, err := mq.NewWriter(exporter.TracertSentTopic)
	if err != nil {
		panic(err)
	}
	tracertWriter = writer

	reader, err := mq.NewReader(exporter.TracertRecvTopic, "tracert_processor_group")
	if err != nil {
		panic(err)
	}
	mq.Subscribe(reader, context.Background(), postTracert)
}

func postTracert(msg *kafka.Message) error {
	var resp exporter.TracertResp
	if err := json.Unmarshal(msg.Value, &resp); err != nil {
		return err
	}
	err := mq.SetTask(mq.Task[map[string]interface{}]{
		ID:       string(msg.Key),
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
	taskID, err := mq.WriteJSONMessages(tracertWriter, exporter.TracertReq{
		Mode:   req.Mode,
		Target: req.Target,
		Port:   uint64(req.Port),
	})
	if err != nil {
		return "", err
	}
	err = mq.SetTask(mq.Task[any]{
		ID:       taskID[0],
		Status:   mq.TaskPending,
		Progress: 0.0,
		Result:   nil,
	})
	if err != nil {
		return "", err
	}
	return taskID[0], nil
}
