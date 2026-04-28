package app

import (
	"LTICore/internal/core/domain"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLaunch_Success(t *testing.T) {
	_, _, _, _, _, platformRepo, loginSessionRepo, _, _ := setupTestServer(t)

	state := "valid-state"
	session := &domain.LoginSession{State: state, Nonce: "nonce123"}
	loginSessionRepo.On("GetAndDelete", mock.Anything, state).Return(session, nil)

	platformRepo.On("GetByIssuerAndClientID", mock.Anything, "https://platform.example", "client123").Return(&domain.Platform{
		JwksUrl:  "https://platform.example/jwks",
		ClientID: "client123",
		Issuer:   "https://platform.example",
	}, nil)

	t.Skip("Launch test requires deeper mocking or refactoring")
}

func TestJWKS_Endpoint(t *testing.T) {
	server, _, _, _, _, _, _, _, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/.well-known/jwks.json", nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var jwks map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &jwks)
	assert.NoError(t, err)
	assert.Contains(t, jwks, "keys")
}
