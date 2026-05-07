package service

import (
	"LTICore/internal/core/domain"
	"context"
	"fmt"
	"log/slog"
)

type AGSMetrics interface {
	IncAGSRequestErrors()
	IncAGSRequestsTotal()
}

type AGSService struct {
	repo    AGSRepository
	metrics AGSMetrics
	log     *slog.Logger
}

func NewAGSService(repo AGSRepository, metrics AGSMetrics) *AGSService {
	return &AGSService{repo: repo, metrics: metrics, log: slog.Default()}
}

func (s *AGSService) GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error) {
	s.log.Info("ags.get_line_items.start", "context_id", contextID)
	s.metrics.IncAGSRequestsTotal()
	if contextID == "" {
		s.metrics.IncAGSRequestErrors()
		s.log.Warn("ags.get_line_items.validation_failed", "reason", "missing_context_id")
		return nil, fmt.Errorf("contextID required")
	}
	items, err := s.repo.GetLineItems(ctx, contextID)
	if err != nil {
		s.metrics.IncAGSRequestErrors()
		s.log.Error("ags.get_line_items.failed", "context_id", contextID, "err", err)
		return nil, err
	}
	s.log.Info("ags.get_line_items.ok", "context_id", contextID, "count", len(items))
	return items, nil
}

func (s *AGSService) CreateLineItem(ctx context.Context, li *domain.LineItem) error {
	s.log.Info("ags.create_line_item.start",
		"context_id", func() string {
			if li != nil {
				return li.ContextID
			}
			return ""
		}(),
		"label", func() string {
			if li != nil {
				return li.Label
			}
			return ""
		}(),
	)
	s.metrics.IncAGSRequestsTotal()
	if li == nil || li.Label == "" || li.ContextID == "" || li.MaxScore <= 0 {
		s.metrics.IncAGSRequestErrors()
		s.log.Warn("ags.create_line_item.validation_failed")
		return fmt.Errorf("invalid lineitem data")
	}
	if err := s.repo.CreateLineItem(ctx, li); err != nil {
		s.metrics.IncAGSRequestErrors()
		s.log.Error("ags.create_line_item.failed", "context_id", li.ContextID, "label", li.Label, "err", err)
		return err
	}
	s.log.Info("ags.create_line_item.ok", "context_id", li.ContextID, "label", li.Label)
	return nil
}

func (s *AGSService) GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error) {
	s.log.Info("ags.get_score.start", "line_item_id", lineItemID, "user_id", userID)
	s.metrics.IncAGSRequestsTotal()
	if lineItemID == "" || userID == "" {
		s.metrics.IncAGSRequestErrors()
		s.log.Warn("ags.get_score.validation_failed")
		return nil, fmt.Errorf("lineItemID and userID required")
	}
	score, err := s.repo.GetScore(ctx, lineItemID, userID)
	if err != nil {
		s.metrics.IncAGSRequestErrors()
		s.log.Error("ags.get_score.failed", "line_item_id", lineItemID, "user_id", userID, "err", err)
		return nil, err
	}
	s.log.Info("ags.get_score.ok", "line_item_id", lineItemID, "user_id", userID)
	return score, nil
}

func (s *AGSService) SaveScore(ctx context.Context, score *domain.Score) error {
	s.log.Info("ags.save_score.start",
		"line_item_id", func() string {
			if score != nil {
				return score.LineItemID
			}
			return ""
		}(),
		"user_id", func() string {
			if score != nil {
				return score.UserID
			}
			return ""
		}(),
	)
	s.metrics.IncAGSRequestsTotal()
	if score == nil || score.LineItemID == "" || score.UserID == "" || score.Score < 0 {
		s.metrics.IncAGSRequestErrors()
		s.log.Warn("ags.save_score.validation_failed")
		return fmt.Errorf("invalid score data")
	}
	if err := s.repo.SaveScore(ctx, score); err != nil {
		s.metrics.IncAGSRequestErrors()
		s.log.Error("ags.save_score.failed", "line_item_id", score.LineItemID, "user_id", score.UserID, "err", err)
		return err
	}
	s.log.Info("ags.save_score.ok", "line_item_id", score.LineItemID, "user_id", score.UserID)
	return nil
}
