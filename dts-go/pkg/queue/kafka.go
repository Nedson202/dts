package queue

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

func NewKafkaClient(brokers []string, topic string) (*KafkaClient, error) {
	log.Printf("Initializing Kafka client with brokers: %v and topic: %s", brokers, topic)
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &KafkaClient{Writer: writer}, nil
}

func (c *KafkaClient) Close() error {
	if err := c.Writer.Close(); err != nil {
		return err
	}
	return c.Reader.Close()
}

func (c *KafkaClient) PublishMessage(ctx context.Context, key, value []byte) error {
	log.Printf("Publishing message to Kafka. Key: %s", string(key))
	err := c.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Printf("Error publishing message to Kafka: %v", err)
		return err
	}
	return nil
}

func (c *KafkaClient) ConsumeMessage(ctx context.Context) ([]byte, error) {
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return msg.Value, nil
}
