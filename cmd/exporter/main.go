package main

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/exporter"
	"VPSBenchmarkBackend/internal/mq"
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Ping(msg *amqp.Delivery) error {
	var req exporter.PingReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	return exporter.Probe(req.Target, req.HostID, req.MonitorType, msg.ReplyTo, msg.MessageId)
}

func Tracert(msg *amqp.Delivery) error {
	var req exporter.TracertReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	return exporter.Tracert(req.Mode, req.Target, req.Port, msg.ReplyTo, msg.MessageId)
}

func main() {
	err := config.Load("config.json")
	if err != nil {
		panic(err)
	}
	cfg := config.Get()
	err = mq.InitMQ(cfg.RabbitMQURL, cfg.RabbitMQPoolSize)
	if err != nil {
		panic(err)
	}
	err = mq.Subscribe(exporter.PingRoute, context.Background(), Ping)
	if err != nil {
		panic(err)
	}
	err = mq.Subscribe(exporter.TracertRoute, context.Background(), Tracert)
	if err != nil {
		panic(err)
	}
	log.Println("Exporter started, waiting for messages...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
