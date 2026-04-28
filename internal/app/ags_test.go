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

func TestAGS_GetLineItems_Success(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	contextID := "context-123"
	expectedItems := []domain.LineItem{
		{ID: "1", Label: "Item 1", MaxScore: 100},
		{ID: "2", Label: "Item 2", MaxScore: 50},
	}

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsRepo.On("GetLineItems", mock.Anything, contextID).Return(expectedItems, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/ags/lineitems?context_id="+contextID, nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		LineItems []domain.LineItem `json:"lineitems"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedItems, resp.LineItems)

	agsMetrics.AssertExpectations(t)
	agsRepo.AssertExpectations(t)
}

func TestAGS_GetLineItems_MissingContext(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsMetrics.On("IncAGSRequestErrors").Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/ags/lineitems", nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	agsRepo.AssertNotCalled(t, "GetLineItems")
	agsMetrics.AssertExpectations(t)
}

func TestAGS_GetLineItems_RepoError(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	contextID := "context-123"
	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsMetrics.On("IncAGSRequestErrors").Once()
	agsRepo.On("GetLineItems", mock.Anything, contextID).Return([]domain.LineItem(nil), errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/ags/lineitems?context_id="+contextID, nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	agsMetrics.AssertExpectations(t)
	agsRepo.AssertExpectations(t)
}

func TestAGS_CreateLineItem_Success(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	newItem := domain.LineItem{
		Label:     "New Item",
		MaxScore:  75,
		ContextID: "context-123",
	}

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsRepo.On("CreateLineItem", mock.Anything, mock.AnythingOfType("*domain.LineItem")).Return(nil)

	body, _ := json.Marshal(newItem)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/ags/lineitems", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	agsMetrics.AssertExpectations(t)
	agsRepo.AssertExpectations(t)
}

func TestAGS_CreateLineItem_InvalidData(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	invalidItem := domain.LineItem{
		Label:     "",
		MaxScore:  75,
		ContextID: "context-123",
	}

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsMetrics.On("IncAGSRequestErrors").Once()

	body, _ := json.Marshal(invalidItem)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/ags/lineitems", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	agsRepo.AssertNotCalled(t, "CreateLineItem")
	agsMetrics.AssertExpectations(t)
}

func TestAGS_GetScore_Success(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	lineItemID := "line-1"
	userID := "user-1"
	expectedScore := &domain.Score{
		UserID:     userID,
		LineItemID: lineItemID,
		Score:      85.5,
	}

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsRepo.On("GetScore", mock.Anything, lineItemID, userID).Return(expectedScore, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/ags/score?lineitem_id="+lineItemID+"&user_id="+userID, nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var score domain.Score
	_ = json.Unmarshal(w.Body.Bytes(), &score)
	assert.Equal(t, expectedScore.Score, score.Score)
	agsMetrics.AssertExpectations(t)
	agsRepo.AssertExpectations(t)
}

func TestAGS_GetScore_MissingParams(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsMetrics.On("IncAGSRequestErrors").Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/ags/score", nil)
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	agsRepo.AssertNotCalled(t, "GetScore")
	agsMetrics.AssertExpectations(t)
}

func TestAGS_SaveScore_Success(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	score := domain.Score{
		UserID:     "user-1",
		LineItemID: "line-1",
		Score:      90,
	}

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsRepo.On("SaveScore", mock.Anything, mock.AnythingOfType("*domain.Score")).Return(nil)

	body, _ := json.Marshal(score)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/ags/score", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	agsMetrics.AssertExpectations(t)
	agsRepo.AssertExpectations(t)
}

func TestAGS_SaveScore_InvalidData(t *testing.T) {
	server, agsRepo, agsMetrics, _, _, _, _, _, _ := setupTestServer(t)

	invalidScore := domain.Score{
		UserID:     "",
		LineItemID: "line-1",
		Score:      90,
	}

	agsMetrics.On("IncAGSRequestsTotal").Once()
	agsMetrics.On("IncAGSRequestErrors").Once()

	body, _ := json.Marshal(invalidScore)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/ags/score", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.api.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	agsRepo.AssertNotCalled(t, "SaveScore")
	agsMetrics.AssertExpectations(t)
}
