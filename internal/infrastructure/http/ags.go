package http

import (
	"LTICore/internal/core/domain"
	http "LTICore/internal/infrastructure/http"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type AGSClient struct {
	client *http.client
}

func NewAGSClient() *AGSClient {
	return &AGSClient{client: NewHTTPClient()}
}

func (c *AGSClient) GetLineItemsFromPlatform(ctx context.Context, agsEndpoint, accessToken string) ([]domain.LineItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", agsEndpoint+"/lineitems", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.ims.lis.v2.lineitemcontainer+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("platform returned %s", resp.Status)
	}

	var items []domain.LineItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Client) SendScoreToPlatform(ctx context.Context, lineItemURL, accessToken string, score *domain.Score) error {
	body, err := json.Marshal(score)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", lineItemURL+"/scores", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/vnd.ims.lis.v2.score+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("platform returned %s", resp.Status)
	}

	return nil
}
