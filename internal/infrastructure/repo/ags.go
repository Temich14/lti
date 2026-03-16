package repo

import (
	"LTICore/internal/core/domain"
	"context"
)

type MockAgsRepo struct {
}

func NewMockAgsRepo() *MockAgsRepo {
	return &MockAgsRepo{}
}

type AGSRepository interface {
	GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error)
	CreateLineItem(ctx context.Context, li *domain.LineItem) error
	GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error)
	SaveScore(ctx context.Context, score *domain.Score) error
}

func (r *MockAgsRepo) GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error) {
	return []domain.LineItem{}, nil
}

func (r *MockAgsRepo) CreateLineItem(ctx context.Context, li *domain.LineItem) error {
	return nil
}

func (r *MockAgsRepo) GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error) {
	return &domain.Score{LineItemID: lineItemID, UserID: userID, Score: 0}, nil
}

func (r *MockAgsRepo) SaveScore(ctx context.Context, score *domain.Score) error {
	return nil
}
