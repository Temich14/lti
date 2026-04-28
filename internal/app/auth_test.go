package app

import (
	"LTICore/internal/core/domain"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	server, _, _, ltiClient, _, platformRepo, _, _, _ := setupTestServer(t)

	openidURL := "https://platform.example/.well-known/openid-configuration"
	regToken := "reg_token"

	ltiClient.On("GetOpenidConfiguration", openidURL).Return(&domain.OpenidConfiguration{
		Issuer:                "https://platform.example",
		RegistrationEndpoint:  "https://platform.example/reg",
		AuthorizationEndpoint: "https://platform.example/auth",
		TokenEndpoint:         "https://platform.example/token",
		JwksUri:               "https://platform.example/jwks",
	}, nil)

	toolRegResponse := &domain.ToolRegistrationResponse{
		ClientId: "client123",
		LtiConfiguration: domain.ToolConfiguration{
			DeploymentID: "deploy456",
		},
	}
	ltiClient.On("SendRegistrationRequest", "https://platform.example/reg", regToken, mock.AnythingOfType("*domain.ToolRegistrationRequest")).Return(toolRegResponse, nil)

	platformRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Platform) bool {
		return p.ClientID == "client123" && p.Issuer == "https://platform.example"
	})).Return(&domain.Platform{ID: uuid.New()}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/auth/register?openid_configuration="+openidURL+"&registration_token="+regToken, nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ltiClient.AssertExpectations(t)
	platformRepo.AssertExpectations(t)
}

func TestRegister_MissingParams(t *testing.T) {
	server, _, _, _, _, _, _, _, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/auth/register", nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_ClientError(t *testing.T) {
	server, _, _, ltiClient, _, platformRepo, _, _, _ := setupTestServer(t)

	openidURL := "https://platform.example/.well-known/openid-configuration"
	regToken := "reg_token"

	ltiClient.On("GetOpenidConfiguration", openidURL).Return((*domain.OpenidConfiguration)(nil), errors.New("network error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/auth/register?openid_configuration="+openidURL+"&registration_token="+regToken, nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ltiClient.AssertExpectations(t)
	platformRepo.AssertNotCalled(t, "Create")
}

func TestLogin_Success(t *testing.T) {
	server, _, _, _, _, platformRepo, loginSessionRepo, _, _ := setupTestServer(t)

	loginReq := domain.LoginRequest{
		Iss:            "https://platform.example",
		ClientID:       "client123",
		LoginHint:      123,
		LTIMessageHint: 456,
	}

	platform := &domain.Platform{
		ClientID: "client123",
		Issuer:   "https://platform.example",
		AuthUrl:  "https://platform.example/auth",
	}
	platformRepo.On("GetByIssuerAndClientID", mock.Anything, loginReq.Iss, loginReq.ClientID).Return(platform, nil)
	loginSessionRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.LoginSession")).Return(nil)

	body, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "https://platform.example/auth")
	platformRepo.AssertExpectations(t)
	loginSessionRepo.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	server, _, _, _, _, _, _, _, _ := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/auth/login", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_PlatformNotFound(t *testing.T) {
	server, _, _, _, _, platformRepo, _, _, _ := setupTestServer(t)

	loginReq := domain.LoginRequest{
		Iss:      "https://unknown.example",
		ClientID: "client123",
	}
	platformRepo.On("GetByIssuerAndClientID", mock.Anything, loginReq.Iss, loginReq.ClientID).Return((*domain.Platform)(nil), errors.New("not found"))

	body, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
