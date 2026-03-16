package domain

import "time"

type LineItem struct {
	ID            string     `json:"id"`
	Label         string     `json:"label"`
	MaxScore      float64    `json:"max_score"`
	ContextID     string     `json:"context_id"`
	StartDateTime *time.Time `json:"start_date_time"`
	EndDateTime   *time.Time `json:"end_date_time"`
}

type Score struct {
	LineItemID string  `json:"lineitem_id"`
	UserID     string  `json:"user_id"`
	Score      float64 `json:"score"`
}

type LineItemFilter struct {
	ResourceID     string `json:"resourceId,omitempty"`
	ResourceLinkID string `json:"resourceLinkId,omitempty"`
	Tag            string `json:"tag,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	Page           int    `json:"page,omitempty"`
}

type ScoreFilter struct {
	Limit     int        `json:"limit,omitempty"`
	Page      int        `json:"page,omitempty"`
	UserID    string     `json:"userId,omitempty"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}
