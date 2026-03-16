package service

import (
	"LTICore/internal/core/domain"
	"context"
	"fmt"
)

type AGSMetrics interface {
	IncAGSRequestErrors()
	IncAGSRequestsTotal()
}

type AGSService struct {
	repo    AGSRepository
	metrics AGSMetrics
}

func NewAGSService(repo AGSRepository, metrics AGSMetrics) *AGSService {
	return &AGSService{repo: repo, metrics: metrics}
}

func (s *AGSService) GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error) {
	s.metrics.IncAGSRequestsTotal()
	if contextID == "" {
		s.metrics.IncAGSRequestErrors()
		return nil, fmt.Errorf("contextID required")
	}
	return s.repo.GetLineItems(ctx, contextID)
}

func (s *AGSService) CreateLineItem(ctx context.Context, li *domain.LineItem) error {
	s.metrics.IncAGSRequestsTotal()
	if li == nil || li.Label == "" || li.ContextID == "" {
		s.metrics.IncAGSRequestErrors()
		return fmt.Errorf("invalid lineitem data")
	}
	return s.repo.CreateLineItem(ctx, li)
}

func (s *AGSService) GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error) {
	s.metrics.IncAGSRequestsTotal()
	if lineItemID == "" || userID == "" {
		s.metrics.IncAGSRequestErrors()
		return nil, fmt.Errorf("lineItemID and userID required")
	}
	return s.repo.GetScore(ctx, lineItemID, userID)
}

func (s *AGSService) SaveScore(ctx context.Context, score *domain.Score) error {
	s.metrics.IncAGSRequestsTotal()
	if score == nil || score.LineItemID == "" || score.UserID == "" {
		s.metrics.IncAGSRequestErrors()
		return fmt.Errorf("invalid score data")
	}
	return s.repo.SaveScore(ctx, score)
}
