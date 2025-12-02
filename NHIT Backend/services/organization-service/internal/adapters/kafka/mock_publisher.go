package kafka

import (
	"context"
	"log"

	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
)

// MockKafkaPublisher is a mock implementation for development/testing
type MockKafkaPublisher struct {
	logger *log.Logger
}

// NewMockKafkaPublisher creates a new mock Kafka publisher
func NewMockKafkaPublisher(logger *log.Logger) ports.KafkaPublisher {
	if logger == nil {
		logger = log.Default()
	}
	return &MockKafkaPublisher{logger: logger}
}

// Publish logs the message instead of actually publishing to Kafka
func (m *MockKafkaPublisher) Publish(ctx context.Context, topic string, message interface{}) error {
	m.logger.Printf("Mock Kafka: Publishing to topic '%s': %+v", topic, message)
	return nil
}

// PublishBatch logs the messages instead of actually publishing to Kafka
func (m *MockKafkaPublisher) PublishBatch(ctx context.Context, topic string, messages []interface{}) error {
	m.logger.Printf("Mock Kafka: Publishing %d messages to topic '%s'", len(messages), topic)
	for i, msg := range messages {
		m.logger.Printf("Message %d: %+v", i+1, msg)
	}
	return nil
}

// Close closes the mock publisher
func (m *MockKafkaPublisher) Close() error {
	m.logger.Println("Mock Kafka: Publisher closed")
	return nil
}
