package queue

import (
	"context"
	"sync"

	"github.com/IBM/sarama"
	"github.com/nedson202/dts-go/pkg/logger"
)

type MessageHandler func([]byte) error

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handler  MessageHandler
	ready    chan bool
}

func NewKafkaConsumer(brokers []string, groupID string, handler MessageHandler) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0 // Use an appropriate Kafka version

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: consumer,
		handler:  handler,
		ready:    make(chan bool),
	}, nil
}

func (kc *KafkaConsumer) Consume(ctx context.Context, topics []string) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := kc.consumer.Consume(ctx, topics, kc); err != nil {
				logger.Error().Err(err).Msg("Error from consumer")
			}
			if ctx.Err() != nil {
				return
			}
			kc.ready = make(chan bool)
		}
	}()

	<-kc.ready // Wait till the consumer has been set up
	logger.Info().Msg("Kafka consumer up and running")

	<-ctx.Done()
	logger.Info().Msg("Terminating: context cancelled")
	wg.Wait()
	return nil
}

func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(kc.ready)
	return nil
}

func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		logger.Info().Msgf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		
		if err := kc.handler(message.Value); err != nil {
			logger.Error().Err(err).Msg("Error handling message")
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func (kc *KafkaConsumer) Close() error {
	return kc.consumer.Close()
}
