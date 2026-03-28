package mq

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func NewWriter(addr, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.Hash{},
	}
}

func NewReader(addr, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{addr},
		Topic:   topic,
		GroupID: groupID,
	})
}

func WriteMessage(writer *kafka.Writer, val []byte) (string, error) {
	u := uuid.New().String()
	err := writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(u),
			Value: val,
		},
	)
	return u, err
}

func WriteJSONMessage(writer *kafka.Writer, v any) (string, error) {
	msg, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return WriteMessage(writer, msg)
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
			err = handler(&m)
			if err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}()
}
