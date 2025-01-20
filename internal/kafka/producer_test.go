package kafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

type mockSyncProducer struct {
	messages []*sarama.ProducerMessage
}

func (m *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	m.messages = append(m.messages, msg)
	return 0, 0, nil
}

func (m *mockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *mockSyncProducer) Close() error {
	return nil
}

func (m *mockSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	return nil
}

func (m *mockSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	return nil
}

func (m *mockSyncProducer) BeginTxn() error {
	return nil
}

func (m *mockSyncProducer) CommitTxn() error {
	return nil
}

func (m *mockSyncProducer) AbortTxn() error {
	return nil
}

func (m *mockSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return 0
}

func (m *mockSyncProducer) IsTransactional() bool {
	return false
}

func TestProducer_SendMessage(t *testing.T) {
	mockProducer := &mockSyncProducer{}
	producer := &SyncProducer{producer: mockProducer}

	err := producer.SendMessage("test-topic", "test-key", []byte("test-value"))
	assert.NoError(t, err)
	assert.Len(t, mockProducer.messages, 1)

	msg := mockProducer.messages[0]
	assert.Equal(t, "test-topic", msg.Topic)
	assert.Equal(t, sarama.StringEncoder("test-key"), msg.Key)
	assert.Equal(t, sarama.ByteEncoder([]byte("test-value")), msg.Value)
}
