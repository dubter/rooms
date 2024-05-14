package kafka

import (
	"app-websocket/internal/config"
	"app-websocket/internal/domain"
	"context"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"log/slog"
)

type ConsumerGroup struct {
	client sarama.ConsumerGroup
	topic  string
	logger *slog.Logger
}

func NewConsumerGroup(cfg *config.KafkaConfig, logger *slog.Logger) (*ConsumerGroup, error) {
	err := pingKafka(cfg.BrokerList, cfg.Topic)
	if err != nil {
		return nil, fmt.Errorf("broker.kafka.NewConsumerGroup: failed to ping Kafka: %w", err)
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.DefaultVersion
	kafkaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup(cfg.BrokerList, cfg.ConsumerGroup, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("broker.kafka.NewConsumerGroup: %w", err)
	}

	return &ConsumerGroup{
		client: client,
		topic:  cfg.Topic,
		logger: logger,
	}, nil
}

func pingKafka(brokerList []string, topic string) error {
	admin, err := sarama.NewClusterAdmin(brokerList, sarama.NewConfig())
	if err != nil {
		return err
	}
	defer admin.Close()

	_, err = admin.DescribeTopics([]string{topic})
	if err != nil {
		return err
	}

	return nil
}

func (cg *ConsumerGroup) Consume(ctx context.Context, handler domain.MessageHandler) error {
	consumer := NewConsumer(handler)
	res := make(chan error)

	go func() {
		for {
			err := cg.client.Consume(ctx, []string{cg.topic}, consumer)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					res <- err
				}

				cg.logger.Error("Failed to consume from kafka", slog.String("error", err.Error()))
			}

		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-cg.client.Errors():
			cg.logger.Error("Failed to write access log entry:", slog.String("error", err.Error()))

		case err := <-res:
			if err != nil {
				return err
			}
		}
	}
}

func (cg *ConsumerGroup) Close() {
	err := cg.client.Close()
	if err != nil {
		cg.logger.Error("broker.kafka.Close", slog.String("error", err.Error()))
	}
}
