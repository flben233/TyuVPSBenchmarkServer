package mq

import (
	"VPSBenchmarkBackend/internal/config"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"log"
	"time"
)

var sharedTransport *kafka.Transport

func NewWriter(topic string) (*kafka.Writer, error) {
	cfg := config.Get()
	mechanism, err := scram.Mechanism(scram.SHA512, cfg.KafkaUser, cfg.KafkaPasswd)
	if err != nil {
		return nil, err
	}
	if sharedTransport == nil {
		sharedTransport = &kafka.Transport{
			SASL: mechanism,
		}
	}
	return &kafka.Writer{
		Addr:      kafka.TCP(cfg.KafkaURL),
		Topic:     topic,
		Balancer:  &kafka.Hash{},
		Transport: sharedTransport,
	}, nil
}

func NewReader(topic, groupID string) (*kafka.Reader, error) {
	cfg := config.Get()
	mechanism, err := scram.Mechanism(scram.SHA512, cfg.KafkaUser, cfg.KafkaPasswd)
	if err != nil {
		return nil, err
	}
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaURL},
		Topic:   topic,
		GroupID: groupID,
		Dialer:  dialer,
	}), nil
}

func WriteMessages(writer *kafka.Writer, val ...[]byte) ([]string, error) {
	keys := make([]string, len(val))
	msgs := make([]kafka.Message, len(val))
	for i, v := range val {
		u := uuid.New().String()
		msgs[i] = kafka.Message{
			Key:   []byte(u),
			Value: v,
		}
		keys[i] = u
	}
	err := writer.WriteMessages(
		context.Background(),
		msgs...,
	)
	return keys, err
}

func WriteJSONMessages(writer *kafka.Writer, v ...any) ([]string, error) {
	msgs := make([][]byte, len(v))
	for i, item := range v {
		msg, err := json.Marshal(item)
		if err != nil {
			return make([]string, 0), err
		}
		msgs[i] = msg
	}
	return WriteMessages(writer, msgs...)
}

func Subscribe(reader *kafka.Reader, ctx context.Context, handler func(message *kafka.Message) error) {
	go func() {
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			go func(m *kafka.Message) {
				err = handler(m)
				if err != nil {
					log.Printf("Error handling message: %v", err)
				}
			}(&m)
		}
	}()
}
