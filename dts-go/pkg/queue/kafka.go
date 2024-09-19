package queue

import (
	"context"

	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

func NewKafkaClient(brokers []string, topic string) (*KafkaClient, error) {
	logger.Info().Msgf("Initializing Kafka client with brokers: %v and topic: %s", brokers, topic)
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
	logger.Info().Msgf("Publishing message to Kafka. Key: %s", string(key))
	err := c.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error publishing message to Kafka")
		return err
	}
	return nil
}

func (c *KafkaClient) ConsumeMessage(ctx context.Context) ([]byte, error) {
	msg, err := c.Reader.ReadMessage(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Error consuming message from Kafka")
		return nil, err
	}
	return msg.Value, nil
}
