package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"LTICore/internal/core/domain"
)

type NRPSClient struct {
	httpClient *http.Client
}

func NewNRPSClient() *NRPSClient {
	return &NRPSClient{httpClient: NewHTTPClient()}
}

func (c *NRPSClient) GetMembers(ctx context.Context, url string, token string) ([]domain.NRPSMember, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := doRequestWithRetry(c.httpClient, req, 5)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NRPS request failed: %s", string(body))
	}

	var nrpsResp struct {
		Members []domain.NRPSMember `json:"members"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&nrpsResp); err != nil {
		return nil, err
	}

	return nrpsResp.Members, nil
}
