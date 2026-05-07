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

// GetMembersPage supports cursor/URL continuation when NRPS implements pagination via a "next" URL or link.
// Some NRPS deployments use different field names, so we parse a few common shapes.
func (c *NRPSClient) GetMembersPage(ctx context.Context, url string, token string) ([]domain.NRPSMember, string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := doRequestWithRetry(c.httpClient, req, 5)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("NRPS request failed: %s", string(body))
	}

	// Minimal flexible decoding of a membership page.
	var page struct {
		Members []domain.NRPSMember `json:"members"`

		Next     string `json:"next"`
		NextPage string `json:"next_page"`
		Links    *struct {
			Next *struct {
				Href string `json:"href"`
			} `json:"next"`
		} `json:"links"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", err
	}

	nextURL := ""
	switch {
	case page.Next != "":
		nextURL = page.Next
	case page.NextPage != "":
		nextURL = page.NextPage
	case page.Links != nil && page.Links.Next != nil:
		nextURL = page.Links.Next.Href
	}

	return page.Members, nextURL, nil
}
