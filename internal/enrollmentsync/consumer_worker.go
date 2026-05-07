package enrollmentsync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type ConsumerService struct {
	repo   *Repository
	reader *kafka.Reader
	logger *slog.Logger
}

func NewConsumerService(repo *Repository, reader *kafka.Reader, logger *slog.Logger) *ConsumerService {
	return &ConsumerService{
		repo:    repo,
		reader:  reader,
		logger:  logger,
	}
}

func (s *ConsumerService) Run(ctx context.Context) {
	for {
		msg, err := s.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.logger.Error("enrollmentsync.consumer.read_failed", "err", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var ev IncomingEvent
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			// Malformed payload: commit to avoid poison-loop (DLQ can be added later).
			s.logger.Error("enrollmentsync.consumer.unmarshal_failed", "err", err, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)
			_ = s.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := s.repo.ApplyUserSynchronizationEvent(ctx, ev); err != nil {
			// Do not commit offsets; Kafka will redeliver and inbox_events will preserve exactly-once effect.
			s.logger.Error("enrollmentsync.consumer.apply_failed", "err", err, "event_id", ev.ID)
			continue
		}

		if err := s.reader.CommitMessages(ctx, msg); err != nil {
			s.logger.Error("enrollmentsync.consumer.commit_failed", "err", err, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)
		}
	}
}

func BuildKafkaReader(brokers []string, topic, groupID string, minBytes, maxBytes int) *kafka.Reader {
	// Note: Commit interval is controlled by CommitMessages() calls.
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		GroupID:  groupID,
		Topic:     topic,
		MinBytes:  minBytes,
		MaxBytes:  maxBytes,
		// Keep MaxWait and ReadTimeout defaults.
		CommitInterval: 0,
	})
}

func TopicValue() string { return fmt.Sprintf("kafka") }

