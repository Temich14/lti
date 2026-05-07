package repo

import (
	"LTICore/internal/core/domain"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

type PostgresAgsRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresAgsRepo(pool *pgxpool.Pool) *PostgresAgsRepo {
	return &PostgresAgsRepo{pool: pool}
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

func (r *PostgresAgsRepo) GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lineitem_id, label, max_score, context_id, start_date_time, end_date_time
		FROM ags_line_items
		WHERE context_id = $1
		ORDER BY created_at DESC
	`, contextID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.LineItem, 0)
	for rows.Next() {
		var (
			id         uuid.UUID
			item       domain.LineItem
			startAt    *time.Time
			endAt      *time.Time
		)
		if err := rows.Scan(&id, &item.Label, &item.MaxScore, &item.ContextID, &startAt, &endAt); err != nil {
			return nil, err
		}
		item.ID = id.String()
		item.StartDateTime = startAt
		item.EndDateTime = endAt
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresAgsRepo) CreateLineItem(ctx context.Context, li *domain.LineItem) error {
	var id uuid.UUID
	if li.ID != "" {
		parsed, err := uuid.Parse(li.ID)
		if err != nil {
			return fmt.Errorf("lineitem id must be uuid: %w", err)
		}
		id = parsed
	} else {
		id = uuid.New()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO ags_line_items (lineitem_id, context_id, label, max_score, start_date_time, end_date_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (lineitem_id) DO UPDATE
		SET context_id = EXCLUDED.context_id,
		    label = EXCLUDED.label,
		    max_score = EXCLUDED.max_score,
		    start_date_time = EXCLUDED.start_date_time,
		    end_date_time = EXCLUDED.end_date_time,
		    updated_at = now()
	`, id, li.ContextID, li.Label, li.MaxScore, li.StartDateTime, li.EndDateTime)
	if err != nil {
		return err
	}
	li.ID = id.String()
	return nil
}

func (r *PostgresAgsRepo) GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error) {
	lineItemUUID, err := uuid.Parse(lineItemID)
	if err != nil {
		return nil, fmt.Errorf("lineitem_id must be uuid: %w", err)
	}

	var score domain.Score
	err = r.pool.QueryRow(ctx, `
		SELECT lineitem_id::text, user_id, score
		FROM ags_scores
		WHERE lineitem_id = $1 AND user_id = $2
	`, lineItemUUID, userID).Scan(&score.LineItemID, &score.UserID, &score.Score)
	if err != nil {
		return nil, err
	}
	return &score, nil
}

func (r *PostgresAgsRepo) SaveScore(ctx context.Context, score *domain.Score) error {
	lineItemUUID, err := uuid.Parse(score.LineItemID)
	if err != nil {
		return fmt.Errorf("lineitem_id must be uuid: %w", err)
	}

	var maxScore float64
	if err := r.pool.QueryRow(ctx, `
		SELECT max_score
		FROM ags_line_items
		WHERE lineitem_id = $1
	`, lineItemUUID).Scan(&maxScore); err != nil {
		return err
	}
	if score.Score < 0 || score.Score > maxScore {
		return fmt.Errorf("score %.2f outside [0, %.2f]", score.Score, maxScore)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO ags_scores (lineitem_id, user_id, score)
		VALUES ($1, $2, $3)
		ON CONFLICT (lineitem_id, user_id) DO UPDATE
		SET score = EXCLUDED.score,
		    updated_at = now()
	`, lineItemUUID, score.UserID, score.Score)
	return err
}
