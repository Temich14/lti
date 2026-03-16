package domain

type LineItem struct {
	ID        string  `json:"id"`
	Label     string  `json:"label"`
	MaxScore  float64 `json:"max_score"`
	ContextID string  `json:"context_id"`
}

type Score struct {
	LineItemID string  `json:"lineitem_id"`
	UserID     string  `json:"user_id"`
	Score      float64 `json:"score"`
}
