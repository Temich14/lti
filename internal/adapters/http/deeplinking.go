package http

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

	sessionID := c.Query("session_id")

	session, err := h.srv.GetDeepLinkingSession(c.Request.Context(), sessionID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid session",
		})
		return
	}

	settings := session.Settings
	platform := session.Platform

	c.HTML(http.StatusOK, "select.html", gin.H{
		"settings":         settings,
		"acceptTypes":      settings.AcceptTypes,
		"acceptMediaTypes": settings.AcceptMediaTypes,
		"acceptMultiple":   settings.AcceptMultiple,
		"title":            settings.Title,
		"text":             settings.Text,
		"data":             settings.Data,
		"returnUrl":        settings.DeepLinkReturnURL,
		"platform":         platform,
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
	typesParam := c.Query("types")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id"})
		return
	}

	session, err := h.srv.GetDeepLinkingSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	_, err = h.srv.GetAvailableContent(
		c.Request.Context(),
		session.ContextID,
		session.Settings.AcceptTypes,
	)

	allowedTypes := make(map[string]bool)
	if typesParam != "" {
		types := strings.Split(typesParam, ",")
		for _, t := range types {
			allowedTypes[t] = true
		}
	}

	content2 := generateTestContent(allowedTypes, session.Settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content2,
		"settings": gin.H{
			"acceptMultiple": session.Settings.AcceptMultiple,
			"autoCreate":     session.Settings.AutoCreate,
		},
	})
}

// generateTestContent генерирует тестовые данные согласно доменным моделям
func generateTestContent(allowedTypes map[string]bool, settings *domain.DeepLinkingSettings) []domain.ContentItem {
	now := time.Now()
	futureTime := now.Add(7 * 24 * time.Hour)
	pastTime := now.Add(-7 * 24 * time.Hour)

	// Базовые тестовые элементы
	allContent := []domain.ContentItem{
		// 1. LTI Resource Link (основной тип)
		{
			Type:  domain.ContentTypeLtiResourceLink,
			URL:   "https://example.com/lti/launch/course-101",
			Title: "Введение в LTI 1.3",
			Text:  "Изучите основные концепции и преимущества LTI 1.3 интеграции с Deep Linking",
			Icon: &domain.Icon{
				URL:    "https://example.com/icons/lti-icon.png",
				Width:  32,
				Height: 32,
			},
			Thumbnail: &domain.Thumbnail{
				URL:    "https://example.com/thumbnails/lti-intro.jpg",
				Width:  200,
				Height: 150,
			},
			WindowTarget: "_blank",
			Custom: map[string]interface{}{
				"course_id":    "CS101",
				"module_id":    "week1",
				"user_role":    "instructor",
				"tool_version": "1.3.0",
				"features":     []string{"deep_linking", "nrps", "ags"},
			},
			// IFrame для встраивания
			IFrame: &domain.IFrame{
				Src:    "https://example.com/embed/lti-intro",
				Width:  800,
				Height: 600,
			},
			// Настройки доступности
			Available: &domain.Availability{
				StartAt: &now,
				EndAt:   &futureTime,
			},
		},

		// 2. LTI Resource Link с заданием (Assignment)
		{
			Type:  domain.ContentTypeLtiResourceLink,
			URL:   "https://example.com/lti/launch/assignment-1",
			Title: "Assignment 1: LTI Integration Project",
			Text:  "Практическое задание по интеграции LTI 1.3 с поддержкой Deep Linking и Assignment Grade Services",
			Icon: &domain.Icon{
				URL:   "https://example.com/icons/assignment-icon.png",
				Width: 32,
			},
			Thumbnail: &domain.Thumbnail{
				URL: "https://example.com/thumbnails/assignment.jpg",
			},
			Custom: map[string]interface{}{
				"assignment_type": "project",
				"max_score":       100,
				"submission_type": "online",
			},
			WindowTarget: "_blank",
			// LineItem для оценивания (AGS)
			LineItem: &domain.LineItem{
				MaxScore:      100.0,
				Label:         "LTI Integration Project",
				ContextID:     "assignment-1",
				StartDateTime: &now,
				EndDateTime:   &futureTime,
			},
			// Настройки отправки
			Submission: &domain.Submission{
				StartAt: &now,
				EndAt:   &futureTime,
			},
			Available: &domain.Availability{
				StartAt: &now,
				EndAt:   &futureTime,
			},
		},

		// 3. Обычная ссылка на внешний ресурс
		{
			Type:         domain.ContentTypeLink,
			URL:          "https://developer.1edtech.org/docs/lti",
			Title:        "LTI Deep Linking Documentation",
			Text:         "Официальная документация по Deep Linking от 1EdTech Consortium",
			WindowTarget: "_blank",
			Thumbnail: &domain.Thumbnail{
				URL:    "https://example.com/thumbnails/docs.jpg",
				Width:  150,
				Height: 150,
			},
		},

		// 4. HTML контент
		{
			Type:  domain.ContentTypeHTML,
			URL:   "https://example.com/html/lesson-1.html",
			Title: "Интерактивный урок: LTI Architecture",
			Text:  "HTML5 урок с интерактивными элементами и практическими заданиями",
			Icon: &domain.Icon{
				URL: "https://example.com/icons/html-icon.png",
			},
			Thumbnail: &domain.Thumbnail{
				URL: "https://example.com/thumbnails/lesson.jpg",
			},
			IFrame: &domain.IFrame{
				Src:    "https://example.com/embed/lesson-1",
				Width:  1024,
				Height: 768,
			},
			WindowTarget: "iframe",
		},

		// 5. Изображение
		{
			Type:  domain.ContentTypeImage,
			URL:   "https://example.com/images/lti-architecture.png",
			Title: "LTI 1.3 Architecture Diagram",
			Text:  "Диаграмма архитектуры LTI 1.3 с указанием всех компонентов и взаимодействий",
			Thumbnail: &domain.Thumbnail{
				URL:    "https://example.com/thumbnails/diagram-thumb.jpg",
				Width:  300,
				Height: 200,
			},
			Icon: &domain.Icon{
				URL: "https://example.com/icons/image-icon.png",
			},
		},

		// 6. Файл (PDF с ограничением по времени)
		{
			Type:  domain.ContentTypeFile,
			URL:   "https://example.com/files/lti-specification.pdf",
			Title: "IMS LTI 1.3 Core Specification",
			Text:  "Полная спецификация IMS LTI 1.3 Core (PDF, 2.4 MB)",
			Icon: &domain.Icon{
				URL: "https://example.com/icons/pdf-icon.png",
			},
			Thumbnail: &domain.Thumbnail{
				URL: "https://example.com/thumbnails/pdf-thumb.jpg",
			},
			// Доступен только в определенный период
			Available: &domain.Availability{
				StartAt: &pastTime,
				EndAt:   &futureTime,
			},
		},

		// 7. LTI Resource Link с квизом и ограничениями по времени
		{
			Type:  domain.ContentTypeLtiResourceLink,
			URL:   "https://example.com/lti/launch/quiz-lti-basics",
			Title: "LTI Knowledge Assessment Quiz",
			Text:  "Проверьте свои знания по LTI 1.3 и Deep Linking. 20 вопросов, 30 минут.",
			Icon: &domain.Icon{
				URL:   "https://example.com/icons/quiz-icon.png",
				Width: 32,
			},
			Custom: map[string]interface{}{
				"quiz_id":       "lti-basics-2024",
				"time_limit":    30,
				"attempts":      2,
				"passing_score": 70,
				"questions":     20,
			},
			LineItem: &domain.LineItem{
				MaxScore:      100.0,
				Label:         "LTI Basics Quiz",
				ContextID:     "quiz-lti-basics",
				StartDateTime: &now,
				EndDateTime:   &futureTime,
			},
			Submission: &domain.Submission{
				StartAt: &now,
				EndAt:   &futureTime,
			},
			Available: &domain.Availability{
				StartAt: &now,
				EndAt:   &futureTime,
			},
			WindowTarget: "_blank",
		},

		// 8. Видео контент (как ссылка)
		{
			Type:  domain.ContentTypeLink,
			URL:   "https://example.com/videos/lti-deep-linking-tutorial.mp4",
			Title: "Video Tutorial: Deep Linking Implementation",
			Text:  "Пошаговое видео-руководство по реализации Deep Linking в LTI 1.3 (45 минут)",
			Thumbnail: &domain.Thumbnail{
				URL:    "https://example.com/thumbnails/video-thumb.jpg",
				Width:  320,
				Height: 180,
			},
			Icon: &domain.Icon{
				URL: "https://example.com/icons/video-icon.png",
			},
			Available: &domain.Availability{
				StartAt: &now,
				EndAt:   &futureTime,
			},
			WindowTarget: "_blank",
		},

		// 9. Инструмент с кастомными параметрами (без заголовка)
		{
			Type: domain.ContentTypeLtiResourceLink,
			URL:  "https://example.com/lti/launch/tool/configurator",
			Text: "Инструмент конфигурации LTI с расширенными настройками",
			Custom: map[string]interface{}{
				"tool_id":     "lti-configurator",
				"mode":        "advanced",
				"permissions": []string{"read", "write", "admin"},
				"settings":    map[string]string{"theme": "dark", "language": "ru", "timezone": "UTC+3"},
				"features":    []string{"deep_linking", "nrps", "ags", "lti_advantage"},
			},
			WindowTarget: "_blank",
		},

		// 10. Ресурс с несколькими вариантами отображения
		{
			Type:  domain.ContentTypeLtiResourceLink,
			URL:   "https://example.com/lti/launch/multiview",
			Title: "Multi-view Learning Resource",
			Text:  "Ресурс с поддержкой различных представлений: iframe, popup, new window",
			IFrame: &domain.IFrame{
				Src:    "https://example.com/embed/multiview",
				Width:  1000,
				Height: 800,
			},
			WindowTarget: "iframe",
			Custom: map[string]interface{}{
				"views":        []string{"iframe", "popup", "new_window"},
				"default_view": "iframe",
			},
		},
	}

	// Фильтруем по разрешенным типам из настроек Deep Linking
	if settings != nil && len(settings.AcceptTypes) > 0 {
		filtered := make([]domain.ContentItem, 0)
		allowedTypesMap := make(map[string]bool)
		for _, t := range settings.AcceptTypes {
			allowedTypesMap[t] = true
		}

		for _, item := range allContent {
			if allowedTypesMap[item.Type] {
				filtered = append(filtered, item)
			}
		}
		return filtered
	}

	// Если есть фильтр по параметру запроса
	if len(allowedTypes) > 0 {
		filtered := make([]domain.ContentItem, 0)
		for _, item := range allContent {
			if allowedTypes[item.Type] {
				filtered = append(filtered, item)
			}
		}
		return filtered
	}

	return allContent
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
