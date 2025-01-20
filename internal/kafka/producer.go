package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

type Producer interface {
	SendMessage(topic string, key string, value []byte) error
	Close() error
}

type SyncProducer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_5_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &SyncProducer{
		producer: producer,
	}, nil
}

func (p *SyncProducer) SendMessage(topic string, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Printf("Message stored in topic(%s)/partition(%d)/offset(%d)", topic, partition, offset)
	return nil
}

func (p *SyncProducer) Close() error {
	return p.producer.Close()
}
