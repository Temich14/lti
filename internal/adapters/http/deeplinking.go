package http

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type DeepLinkingHandler struct {
	srv DeepLinkingService
}

type DeepLinkingService interface {
	HandleDeepLinkingRequest(ctx context.Context, idToken string) (*domain.DeepLinkingSettings, error)
	BuildDeepLinkResponse(ctx context.Context, settings *domain.DeepLinkingSettings, items []domain.ContentItem, platform *domain.Platform) (string, error)
	CreateContentItemForLtiResource(title, text, url string, lineItem *domain.LineItem) *domain.ContentItem
	StoreDeepLinkingSession(ctx context.Context, sessionID string, session *domain.DeepLinkingSession) error
	GetDeepLinkingSession(ctx context.Context, sessionID string) (*domain.DeepLinkingSession, error)
	DeleteDeepLinkingSession(ctx context.Context, sessionID string) error
	GetAvailableContent(ctx context.Context, contextID string, acceptTypes []string) ([]domain.ContentItem, error)
	CreateLineItem(ctx context.Context, lineItem *domain.LineItem) error
}

func NewDeepLinkingHandler(srv DeepLinkingService) *DeepLinkingHandler {
	return &DeepLinkingHandler{srv: srv}
}

// ShowContentSelection отображает страницу выбора контента для Deep Linking
func (h *DeepLinkingHandler) ShowContentSelection(c *gin.Context) {
	// Получаем JWT из формы (LTI launch request)
	idToken := c.PostForm("id_token")
	if idToken == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Missing id_token",
		})
		return
	}

	// Извлекаем настройки Deep Linking из токена
	settings, err := h.srv.HandleDeepLinkingRequest(c.Request.Context(), idToken)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid deep linking request: " + err.Error(),
		})
		return
	}

	// Получаем платформу из контекста (должна быть установлена в middleware)
	platform, exists := c.Get("platform")
	if !exists {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Platform not found",
		})
		return
	}

	// Сохраняем настройки и платформу в сессии для последующего использования
	sessionID := generateSessionID()
	err = h.srv.StoreDeepLinkingSession(c.Request.Context(), sessionID, &domain.DeepLinkingSession{
		Settings:  settings,
		Platform:  platform.(*domain.Platform),
		UserID:    c.GetString("user_id"),
		ContextID: c.GetString("context_id"),
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to create session",
		})
		return
	}

	// Устанавливаем cookie с ID сессии
	c.SetCookie("dl_session", sessionID, 3600, "/", "", false, true)

	// Рендерим страницу выбора контента
	c.HTML(http.StatusOK, "deeplink/select.html", gin.H{
		"settings":         settings,
		"acceptTypes":      settings.AcceptTypes,
		"acceptMediaTypes": settings.AcceptMediaTypes,
		"acceptMultiple":   settings.AcceptMultiple,
		"title":            settings.Title,
		"text":             settings.Text,
		"data":             settings.Data,
		"returnUrl":        settings.DeepLinkReturnURL,
	})
}

// ReturnContent обрабатывает возврат выбранного контента в платформу
func (h *DeepLinkingHandler) ReturnContent(c *gin.Context) {
	// Получаем сессию из cookie
	sessionID, err := c.Cookie("dl_session")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session"})
		return
	}

	// Парсим выбранные элементы из запроса
	var request struct {
		Items []struct {
			Type       string `json:"type" binding:"required"`
			Title      string `json:"title"`
			Text       string `json:"text"`
			URL        string `json:"url"`
			ResourceID string `json:"resource_id"`
			LineItem   *struct {
				ScoreMaximum float64 `json:"scoreMaximum"`
				Label        string  `json:"label"`
			} `json:"lineItem,omitempty"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем данные сессии
	session, err := h.srv.GetDeepLinkingSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session"})
		return
	}

	// Конвертируем в ContentItems
	var contentItems []domain.ContentItem
	for _, item := range request.Items {
		var lineItem *domain.LineItem
		if item.LineItem != nil {
			lineItem = &domain.LineItem{
				Label:     item.LineItem.Label,
				MaxScore:  item.LineItem.ScoreMaximum,
				ContextID: session.ContextID,
			}
		}

		contentItem := h.srv.CreateContentItemForLtiResource(
			item.Title,
			item.Text,
			item.URL,
			lineItem,
		)
		contentItems = append(contentItems, *contentItem)
	}

	// Строим JWT ответ
	jwtResponse, err := h.srv.BuildDeepLinkResponse(
		c.Request.Context(),
		session.Settings,
		contentItems,
		session.Platform,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build response"})
		return
	}

	// Очищаем сессию
	h.srv.DeleteDeepLinkingSession(c.Request.Context(), sessionID)
	c.SetCookie("dl_session", "", -1, "/", "", false, true)

	// Если запрос пришёл как JSON (AJAX из select.html) — возвращаем JSON.
	accept := c.GetHeader("Accept")
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(accept, "application/json") {
		c.JSON(http.StatusOK, gin.H{
			"jwt":       jwtResponse,
			"returnUrl": session.Settings.DeepLinkReturnURL,
		})
		return
	}

	// Иначе — HTML с формой для отправки JWT обратно в платформу
	c.HTML(http.StatusOK, "deeplink/return.html", gin.H{
		"returnUrl": session.Settings.DeepLinkReturnURL,
		"jwt":       jwtResponse,
	})
}

// APIReturnContent API версия для возврата контента (без HTML)
func (h *DeepLinkingHandler) APIReturnContent(c *gin.Context) {
	var request struct {
		JWT string `json:"jwt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"jwt":    request.JWT,
	})
}

func (h *DeepLinkingHandler) GetAvailableContent(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id"})
		return
	}

	session, err := h.srv.GetDeepLinkingSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	content, err := h.srv.GetAvailableContent(
		c.Request.Context(),
		session.ContextID,
		session.Settings.AcceptTypes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"settings": gin.H{
			"acceptMultiple": session.Settings.AcceptMultiple,
			"autoCreate":     session.Settings.AutoCreate,
		},
	})
}

func (h *DeepLinkingHandler) CreateLineItemFromSelection(c *gin.Context) {
	var request struct {
		ResourceID   string  `json:"resource_id" binding:"required"`
		Label        string  `json:"label" binding:"required"`
		ScoreMaximum float64 `json:"scoreMaximum" binding:"required,gt=0"`
		ContextID    string  `json:"context_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lineItem := &domain.LineItem{
		Label:     request.Label,
		MaxScore:  request.ScoreMaximum,
		ContextID: request.ContextID,
	}

	if err := h.srv.CreateLineItem(c.Request.Context(), lineItem); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lineItem)
}

func (h *DeepLinkingHandler) CancelDeepLinking(c *gin.Context) {
	sessionID, err := c.Cookie("dl_session")
	if err == nil {
		h.srv.DeleteDeepLinkingSession(c.Request.Context(), sessionID)
		c.SetCookie("dl_session", "", -1, "/", "", false, true)
	}

	c.HTML(http.StatusOK, "deeplink/return.html", gin.H{
		"returnUrl": c.Query("return_url"),
		"jwt":       "",
		"error":     "User cancelled selection",
	})
}

func generateSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
