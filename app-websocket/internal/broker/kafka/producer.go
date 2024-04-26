package kafka

import (
	"app-websocket/internal/config"
	"app-websocket/internal/domain"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log/slog"
	"time"
)

type KafkaProducer struct {
	client sarama.AsyncProducer
	topic  string
	logger *slog.Logger
}

func NewProducer(cfg *config.KafkaConfig, logger *slog.Logger) (*KafkaProducer, error) {
	err := pingKafka(cfg.BrokerList, cfg.Topic)
	if err != nil {
		return nil, fmt.Errorf("broker.kafka.NewProducer: failed to ping Kafka: %w", err)
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.DefaultVersion
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	kafkaConfig.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	client, err := sarama.NewAsyncProducer(cfg.BrokerList, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("broker.kafka.New: %w", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range client.Errors() {
			logger.Error("producer error:", slog.String("error", err.Error()))
		}
	}()

	return &KafkaProducer{
		client: client,
		topic:  cfg.Topic,
		logger: logger,
	}, nil
}

func (kp *KafkaProducer) Produce(msg *domain.Message) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("broker.kafka.Produce: %w", err)
	}

	kp.client.Input() <- &sarama.ProducerMessage{
		Topic: kp.topic,
		Key:   sarama.ByteEncoder(msg.RoomID),
		Value: sarama.ByteEncoder(jsonMsg),
	}

	return nil
}

func (kp *KafkaProducer) Close() {
	err := kp.client.Close()
	if err != nil {
		kp.logger.Error("broker.kafka.Close", slog.String("error", err.Error()))
	}
}
