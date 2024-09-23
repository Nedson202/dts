package queue

import (
	"context"
	"sync"

	"github.com/nedson202/dts-go/pkg/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaClient struct {
	client   *kgo.Client
	messages chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewKafkaClient(brokers []string, groupID string, topic string) (*KafkaClient, error) {
	logger.Info().Msgf("Starting Kafka client for topic: %v", topic)
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Error().Err(err).Msgf("Error creating Kafka client: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaClient{
		client:   client,
		messages: make(chan []byte),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (kc *KafkaClient) Consume() error {
	kc.wg.Add(1)
	go func() {
		defer kc.wg.Done()
		for {
			logger.Info().Msgf("Polling for messages...")
			fetches := kc.client.PollFetches(kc.ctx)
			if fetches.IsClientClosed() {
				logger.Info().Msgf("Kafka client is closed")
				close(kc.messages)
				return
			}
			fetches.EachRecord(func(record *kgo.Record) {
				logger.Info().Msgf("Received message: %s", string(record.Value))
				kc.messages <- record.Value
			})
		}
	}()

	return nil
}

func (kc *KafkaClient) Messages() <-chan []byte {
	return kc.messages
}

func (kc *KafkaClient) Produce(ctx context.Context, topic string, key, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	return kc.client.ProduceSync(ctx, record).FirstErr()
}

func (kc *KafkaClient) Close() error {
	logger.Info().Msgf("Closing Kafka client")
	kc.cancel() // Cancel the context to stop all goroutines
	kc.wg.Wait() // Wait for all goroutines to finish
	kc.client.Close()
	close(kc.messages)
	return nil
}
