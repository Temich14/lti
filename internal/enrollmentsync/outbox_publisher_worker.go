package enrollmentsync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type OutboxPublisherService struct {
	repo *Repository

	writer       *kafka.Writer
	pollInterval time.Duration
	batchSize    int
	lockLease    time.Duration
	maxAttempts  int32

	logger *slog.Logger
}

func NewOutboxPublisherService(repo *Repository, writer *kafka.Writer, pollInterval time.Duration, batchSize int, lockLease time.Duration, maxAttempts int32, logger *slog.Logger) *OutboxPublisherService {
	return &OutboxPublisherService{
		repo:          repo,
		writer:        writer,
		pollInterval: pollInterval,
		batchSize:     batchSize,
		lockLease:     lockLease,
		maxAttempts:   maxAttempts,
		logger:        logger,
	}
}

func (s *OutboxPublisherService) RunKafkaOutboxPublisher(ctx context.Context, owner string) {
	if owner == "" {
		owner = fmt.Sprintf("outbox-publisher-%d", time.Now().UnixNano())
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		events, err := s.repo.ClaimOutboxEvents(ctx, owner, s.batchSize, s.lockLease)
		if err != nil {
			s.logger.Error("enrollmentsync.outbox.claim_failed", "err", err)
			continue
		}
		if len(events) == 0 {
			continue
		}

		for _, ev := range events {
			// If we already exceeded max attempts, treat it as poison and stop retrying.
			if ev.Attempts >= s.maxAttempts {
				_ = s.repo.MarkOutboxDeadLetter(ctx, owner, ev.ID, "max attempts exceeded")
				continue
			}

			msg := kafka.Message{
				Topic: s.writer.Topic,
				Key:   []byte(ev.PartitionKey),
				Value: ev.Envelope,
			}

			writeErr := s.writer.WriteMessages(ctx, msg)
			if writeErr != nil {
				s.logger.Error("enrollmentsync.outbox.publish_failed", "event_id", ev.EventID, "err", writeErr)
				_ = s.repo.MarkOutboxFailed(ctx, owner, ev.ID, writeErr.Error())
				continue
			}

			if err := s.repo.MarkOutboxProcessed(ctx, owner, ev.ID); err != nil {
				s.logger.Error("enrollmentsync.outbox.mark_processed_failed", "event_id", ev.EventID, "err", err)
			}
		}
	}
}

