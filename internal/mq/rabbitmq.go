package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type subscriber struct {
	routingKey string
	ctx        context.Context
	handler    func(d *amqp.Delivery) error
}

var (
	sharedConn        *amqp.Connection
	sharedChannel     chan *amqp.Channel
	mqURL             string
	mqPoolSize        int
	lateSubscriptions = make([]*subscriber, 0)
)

func getConn() (*amqp.Connection, error) {
	if sharedConn != nil {
		return sharedConn, nil
	}
	conn, err := amqp.Dial(mqURL)
	if err != nil {
		return nil, err
	}
	sharedConn = conn
	errCh := make(chan *amqp.Error, 1)
	conn.NotifyClose(errCh)
	go func() {
		for err := range errCh {
			log.Printf("RabbitMQ connection closed: %v. Attempting to reconnect...", err)
			time.Sleep(1 * time.Second) // backoff before reconnecting
			newConn, err := amqp.Dial(mqURL)
			if err != nil {
				log.Printf("Failed to reconnect to RabbitMQ: %v", err)
				continue
			}
			sharedConn = newConn
			newConn.NotifyClose(errCh)
			sharedChannel = nil // reset channel pool
			log.Println("Successfully reconnected to RabbitMQ")
		}
	}()
	return conn, nil
}

func getChannel() (*amqp.Channel, error) {
	if sharedChannel == nil {
		sharedChannel = make(chan *amqp.Channel, mqPoolSize)
		for i := 0; i < mqPoolSize; i++ {
			ch, err := getConn()
			if err != nil {
				return nil, err
			}
			channel, err := ch.Channel()
			if err != nil {
				return nil, err
			}
			sharedChannel <- channel
		}
	}
	return <-sharedChannel, nil
}

func returnChannel(ch *amqp.Channel) {
	if sharedChannel == nil {
		ch.Close()
		return
	}
	if ch.IsClosed() {
		ch, err := newChannel("")
		if err != nil {
			log.Printf("Failed to create new RabbitMQ channel: %v", err)
			return
		}
		sharedChannel <- ch
	} else {
		sharedChannel <- ch
	}
}

func newChannel(routingKey string) (*amqp.Channel, error) {
	conn, err := getConn()
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if routingKey != "" {
		_, err = ch.QueueDeclare(
			routingKey,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			ch.Close()
			return nil, err
		}
	}
	return ch, nil
}

func InitMQ(url string, poolSize int) error {
	mqURL = url
	mqPoolSize = poolSize
	for _, sub := range lateSubscriptions {
		err := Subscribe(sub.routingKey, sub.ctx, sub.handler)
		if err != nil {
			return fmt.Errorf("failed to subscribe to RabbitMQ topic %s: %v", sub.routingKey, err)
		} else {
			log.Printf("Successfully subscribed to RabbitMQ topic %s", sub.routingKey)
		}
	}
	lateSubscriptions = nil
	return nil
}

func PublishJSON(routingKey, replyTo string, v any) error {
	_, err := PublishJSONWithID(routingKey, replyTo, v, "")
	return err
}

func PublishJSONWithID(routingKey, replyTo string, v any, id string) (string, error) {
	msg, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if id == "" {
		id = uuid.New().String()
	}
	ch, err := getChannel()
	if err != nil {
		return "", err
	}
	defer returnChannel(ch)
	return id, ch.Publish(
		"",         // exchange
		routingKey, // routing key
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msg,
			ReplyTo:      replyTo,
			MessageId:    id,
		},
	)
}

func LateSubscribe(routingKey string, ctx context.Context, handler func(d *amqp.Delivery) error) {
	lateSubscriptions = append(lateSubscriptions, &subscriber{
		routingKey: routingKey,
		ctx:        ctx,
		handler:    handler,
	})
}

func Subscribe(routingKey string, ctx context.Context, handler func(d *amqp.Delivery) error) error {
	initCh := func() (*amqp.Channel, <-chan amqp.Delivery, <-chan *amqp.Error, error) {
		ch, err := newChannel(routingKey)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create RabbitMQ channel: %w", err)
		}
		msgs, err := ch.Consume(
			routingKey,
			"",    // consumer
			false, // autoAck
			false, // exclusive
			false, // noLocal
			false, // noWait
			nil,   // args
		)
		if err != nil {
			ch.Close()
			return nil, nil, nil, fmt.Errorf("failed to consume from RabbitMQ queue: %w", err)
		}
		closeErr := make(chan *amqp.Error, 1)
		ch.NotifyClose(closeErr)
		return ch, msgs, closeErr, nil
	}
	ch, msgs, errCh, err := initCh()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				ch.Close()
				return
			case d := <-msgs:
				go func() {
					err := handler(&d)
					if err != nil {
						log.Printf("Error handling RabbitMQ message: %v", err)
						d.Nack(false, false) // requeue on error
					} else {
						d.Ack(false)
					}
				}()
			case err := <-errCh:
				log.Printf("RabbitMQ channel closed: %v. Attempting to reconnect...", err)
				time.Sleep(1 * time.Second) // backoff before reconnecting
				ch, msgs, errCh, _ = initCh()
			}
		}
	}()
	return nil
}
