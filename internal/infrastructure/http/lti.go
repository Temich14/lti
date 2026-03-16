package http

import (
	"LTICore/internal/core/domain"
	bytes2 "bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

type LTIClient struct {
	client *http.Client
	log    *slog.Logger
}

func NewLtiClient(log *slog.Logger) *LTIClient {
	return &LTIClient{
		client: NewHTTPClient(),
		log:    log,
	}
}

func (c *LTIClient) GetOpenidConfiguration(openidUrl string) (*domain.OpenidConfiguration, error) {
	req := httptest.NewRequest("GET", openidUrl, nil)
	resp, err := doRequestWithRetry(c.client, req, 3)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var openidConfiguration domain.OpenidConfiguration
	err = json.Unmarshal(body, &openidConfiguration)

	if err != nil {
		return nil, err
	}
	return &openidConfiguration, nil
}

func (c *LTIClient) GetAccessToken(ctx context.Context, platform *domain.Platform, scope, privateKeyID string, key *rsa.PrivateKey) (string, error) {

	claims := jwt.MapClaims{
		"iss": platform.ClientID,
		"sub": platform.ClientID,
		"aud": platform.TokenUrl,
		"iat": jwt.NewNumericDate(jwt.TimeFunc()),
		"exp": jwt.NewNumericDate(jwt.TimeFunc().Add(5 * 60 * 1e9)),
		"jti": "random-id",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = privateKeyID

	signedJWT, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"grant_type":            "client_credentials",
		"scope":                 scope,
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		"client_assertion":      signedJWT,
	}

	body := make(url.Values)
	for k, v := range data {
		body.Set(k, v)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		platform.TokenUrl,
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := doRequestWithRetry(c.client, req, 5)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (c *LTIClient) SendRegistrationRequest(registrationUrl, registrationToken string, body *domain.ToolRegistrationRequest) (*domain.ToolRegistrationResponse, error) {
	req, err := prepareRegistrationRequest(registrationUrl, registrationToken, body)
	if err != nil {
		return nil, err
	}
	resp, err := doRequestWithRetry(c.client, req, 5)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var platform domain.ToolRegistrationResponse
	err = json.NewDecoder(resp.Body).Decode(&platform)
	if err != nil {
		c.log.Error("Error decoding registration response", "error", err)
		return nil, err
	}
	return &platform, nil
}

func prepareRegistrationRequest(registrationUrl, registrationToken string, body *domain.ToolRegistrationRequest) (*http.Request, error) {
	u, _ := url.Parse(registrationUrl)
	q := u.Query()
	q.Add("registration_token", registrationToken)
	u.RawQuery = q.Encode()

	byteBody, err := json.Marshal(body)

	req, err := http.NewRequest("POST", u.String(), bytes2.NewBuffer(byteBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+registrationToken)

	return req, nil
}
