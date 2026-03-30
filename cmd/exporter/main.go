package main

import (
	"VPSBenchmarkBackend/internal/exporter"
	"VPSBenchmarkBackend/internal/mq"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"os"
	"os/signal"
	"syscall"
)

const (
	pingGroup    = "exporter_ping"
	tracertGroup = "exporter_tracert"
)

func Ping(msg *kafka.Message) error {
	var req exporter.PingReq
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return err
	}
	return exporter.Ping(req.Target, req.HostID)
}

func Tracert(msg *kafka.Message) error {
	var req exporter.TracertReq
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return err
	}
	return exporter.Tracert(req.Mode, req.Target, req.Port)
}

func main() {
	pingReader, err := mq.NewReader(exporter.PingSentTopic, pingGroup)
	if err != nil {
		panic(err)
	}
	mq.Subscribe(pingReader, context.Background(), Ping)

	tracertReader, err := mq.NewReader(exporter.TracertSentTopic, tracertGroup)
	if err != nil {
		panic(err)
	}
	mq.Subscribe(tracertReader, context.Background(), Tracert)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	pingReader.Close()
	tracertReader.Close()
}
