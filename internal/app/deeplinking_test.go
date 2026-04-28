package app

import (
	"LTICore/internal/core/domain"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeepLinking_ShowContentSelection_Success(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	idToken := "valid.id.token"
	settings := &domain.DeepLinkingSettings{
		AcceptTypes:       []string{"ltiResourceLink"},
		AcceptMultiple:    true,
		DeepLinkReturnURL: "https://platform.example/return",
		Title:             "Select Content",
		Text:              "Choose an item",
		Data:              "some-data",
	}

	dlService.On("HandleDeepLinkingRequest", mock.Anything, idToken).Return(settings, nil)
	dlService.On("StoreDeepLinkingSession", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.DeepLinkingSession")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/select", bytes.NewReader([]byte("id_token="+idToken)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	dlService.AssertExpectations(t)
}

func TestDeepLinking_ShowContentSelection_MissingToken(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/select", nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	dlService.AssertNotCalled(t, "HandleDeepLinkingRequest")
}

func TestDeepLinking_ShowContentSelection_InvalidToken(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	idToken := "invalid"
	dlService.On("HandleDeepLinkingRequest", mock.Anything, idToken).Return((*domain.DeepLinkingSettings)(nil), errors.New("invalid token"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/select", bytes.NewReader([]byte("id_token="+idToken)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	dlService.AssertExpectations(t)
}

func TestDeepLinking_ReturnContent_Success(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	sessionID := "sess123"
	settings := &domain.DeepLinkingSettings{
		DeepLinkReturnURL: "https://platform.example/return",
		Data:              "some-data",
	}
	platform := &domain.Platform{ClientID: "client123"}
	session := &domain.DeepLinkingSession{Settings: settings, Platform: platform, ContextID: "context-123"}

	requestBody := map[string]any{
		"items": []map[string]any{
			{"type": "ltiResourceLink", "title": "Test Resource", "url": "https://tool.example/resource/1"},
		},
	}
	body, _ := json.Marshal(requestBody)

	dlService.On("GetDeepLinkingSession", mock.Anything, sessionID).Return(session, nil)
	dlService.On("CreateContentItemForLtiResource", "Test Resource", "", "https://tool.example/resource/1", (*domain.LineItem)(nil)).
		Return(&domain.ContentItem{Type: "ltiResourceLink", URL: "https://tool.example/resource/1", Title: "Test Resource"})
	dlService.On("BuildDeepLinkResponse", mock.Anything, settings, mock.AnythingOfType("[]domain.ContentItem"), platform).Return("signed-jwt", nil)
	dlService.On("DeleteDeepLinkingSession", mock.Anything, sessionID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/return", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "dl_session", Value: sessionID})
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	dlService.AssertExpectations(t)
}

func TestDeepLinking_ReturnContent_MissingSessionCookie(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	body, _ := json.Marshal(map[string]any{"items": []any{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/return", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	dlService.AssertNotCalled(t, "GetDeepLinkingSession")
}

func TestDeepLinking_ReturnContent_InvalidItems(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	sessionID := "sess123"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/return", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "dl_session", Value: sessionID})
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	dlService.AssertNotCalled(t, "GetDeepLinkingSession")
}

func TestDeepLinking_APIReturnContent_Success(t *testing.T) {
	server, _, _, _, _, _, _, _, _ := setupTestServer(t)

	body, _ := json.Marshal(map[string]string{"jwt": "signed.jwt.token"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/api/return", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "success", resp["status"])
	assert.Equal(t, "signed.jwt.token", resp["jwt"])
}

func TestDeepLinking_GetAvailableContent_Success(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	sessionID := "sess123"
	session := &domain.DeepLinkingSession{
		Settings:  &domain.DeepLinkingSettings{AcceptTypes: []string{"ltiResourceLink"}},
		ContextID: "context-123",
	}
	content := []domain.ContentItem{
		{Type: "ltiResourceLink", Title: "Resource 1"},
		{Type: "ltiResourceLink", Title: "Resource 2"},
	}

	dlService.On("GetDeepLinkingSession", mock.Anything, sessionID).Return(session, nil)
	dlService.On("GetAvailableContent", mock.Anything, "context-123", []string{"ltiResourceLink"}).Return(content, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/deeplink/content?session_id="+sessionID, nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	dlService.AssertExpectations(t)
}

func TestDeepLinking_GetAvailableContent_NoSession(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	dlService.On("GetDeepLinkingSession", mock.Anything, "invalid").Return((*domain.DeepLinkingSession)(nil), errors.New("not found"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/deeplink/content?session_id=invalid", nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	dlService.AssertExpectations(t)
}

func TestDeepLinking_CreateLineItemFromSelection_Success(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	body, _ := json.Marshal(map[string]any{
		"resource_id":  "res-123",
		"label":        "Assignment",
		"scoreMaximum": 100.0,
		"context_id":   "context-456",
	})

	dlService.On("CreateLineItem", mock.Anything, mock.MatchedBy(func(li *domain.LineItem) bool {
		return li.Label == "Assignment" && li.MaxScore == 100.0 && li.ContextID == "context-456"
	})).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/lineitem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	dlService.AssertExpectations(t)
}

func TestDeepLinking_CreateLineItemFromSelection_InvalidData(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	body, _ := json.Marshal(map[string]any{
		"label":        "Assignment",
		"scoreMaximum": 100.0,
		"context_id":   "context-456",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/deeplink/lineitem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	dlService.AssertNotCalled(t, "CreateLineItem")
}

func TestDeepLinking_CancelDeepLinking(t *testing.T) {
	server, _, _, _, _, _, _, _, dlService := setupTestServer(t)

	sessionID := "sess123"
	returnURL := "https://platform.example/return"

	dlService.On("DeleteDeepLinkingSession", mock.Anything, sessionID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/deeplink/cancel?return_url="+returnURL, nil)
	req.AddCookie(&http.Cookie{Name: "dl_session", Value: sessionID})
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	dlService.AssertExpectations(t)
}
