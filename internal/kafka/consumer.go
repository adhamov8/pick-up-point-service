package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type ConsumerHandler interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error
}

func NewConsumer(brokers []string, groupID string, topics []string, handler ConsumerHandler) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0

	ctx := context.Background()
	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return err
	}

	go func() {
		for {
			if err := client.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}
