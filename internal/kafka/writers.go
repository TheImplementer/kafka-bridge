package kafka

import (
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

// WriterPool lazily creates writers per topic and reuses them.
type WriterPool struct {
	mu      sync.Mutex
	writers map[string]*kafka.Writer
	brokers []string
	client  string
}

// NewWriterPool builds a writer pool for the provided brokers and client ID.
func NewWriterPool(brokers []string, clientID string) *WriterPool {
	return newWriterPool(brokers, clientID)
}

func newWriterPool(brokers []string, clientID string) *WriterPool {
	return &WriterPool{
		brokers: brokers,
		client:  clientID,
		writers: make(map[string]*kafka.Writer),
	}
}

// Get returns a writer bound to the destination topic.
func (p *WriterPool) Get(topic string) *kafka.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if writer, ok := p.writers[topic]; ok {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	if p.client != "" {
		writer.Transport = &kafka.Transport{
			ClientID: p.client,
		}
	}

	p.writers[topic] = writer
	return writer
}

// Close flushes and closes all managed writers.
func (p *WriterPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close writer %s: %w", topic, err)
		}
	}
	return firstErr
}
