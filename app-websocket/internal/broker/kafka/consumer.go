package kafka

import (
	"app-websocket/internal/domain"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type Consumer struct {
	handler domain.MessageHandler
}

func NewConsumer(handler domain.MessageHandler) *Consumer {
	return &Consumer{
		handler: handler,
	}
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return fmt.Errorf("broker.kafka.ConsumeClaim: Messages channel is closed")
			}

			var domainMsg domain.Message
			err := json.Unmarshal(msg.Value, &domainMsg)
			if err != nil {
				return fmt.Errorf("broker.kafka.ConsumeClaim: %w", err)
			}

			err = c.handler(domainMsg)
			if err != nil {
				return fmt.Errorf("broker.kafka.ConsumeClaim: %w", err)
			}

			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			session.Commit()
			return nil
		}
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
