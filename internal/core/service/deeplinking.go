package service

import (
	"LTICore/internal/config"
	"LTICore/internal/core/domain"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type DeepLinkingService struct {
	ltiService *LTIService
	cfg        *config.Config
	log        *slog.Logger
	mu         sync.RWMutex
	sessions   map[string]*domain.DeepLinkingSession
}

func NewDeepLinkingService(ltiService *LTIService, cfg *config.Config) *DeepLinkingService {
	return &DeepLinkingService{
		ltiService: ltiService,
		cfg:        cfg,
		log:        slog.Default(),
		sessions:   make(map[string]*domain.DeepLinkingSession),
	}
}

func (s *DeepLinkingService) HandleDeepLinkingRequest(ctx context.Context, idToken string) (*domain.DeepLinkingSettings, error) {
	s.log.Info("deeplink.handle_request.start")
	token, err := s.ltiService.parseUnverifiedToken(idToken)
	if err != nil {
		s.log.Error("deeplink.handle_request.parse_token.failed", "err", err)
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)

	dlClaim, ok := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"]
	if !ok {
		s.log.Warn("deeplink.handle_request.missing_settings")
		return nil, fmt.Errorf("missing deep linking settings")
	}

	dlJSON, _ := json.Marshal(dlClaim)
	var settings domain.DeepLinkingSettings
	if err := json.Unmarshal(dlJSON, &settings); err != nil {
		s.log.Error("deeplink.handle_request.unmarshal.failed", "err", err)
		return nil, fmt.Errorf("invalid deep linking settings: %w", err)
	}

	s.log.Info("deeplink.handle_request.ok", "accept_multiple", settings.AcceptMultiple, "accept_types_count", len(settings.AcceptTypes))
	return &settings, nil
}

func (s *DeepLinkingService) BuildDeepLinkResponse(
	ctx context.Context,
	settings *domain.DeepLinkingSettings,
	items []domain.ContentItem,
	platform *domain.Platform,
) (string, error) {

	s.log.Info("deeplink.build_response.start", "client_id", platform.ClientID, "item_count", len(items))
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   s.cfg.LTIConfig.Domain,
		"aud":   []string{platform.ClientID},
		"exp":   now.Add(5 * time.Minute).Unix(),
		"iat":   now.Unix(),
		"nonce": generateRandomString(32),
		"https://purl.imsglobal.org/spec/lti-dl/claim/content_items": items,
		"https://purl.imsglobal.org/spec/lti-dl/claim/data":          settings.Data,
		"https://purl.imsglobal.org/spec/lti-dl/claim/msg":           "success",
		"https://purl.imsglobal.org/spec/lti-dl/claim/errormsg":      "",
		"https://purl.imsglobal.org/spec/lti-dl/claim/log":           "",
		"https://purl.imsglobal.org/spec/lti-dl/claim/errorlog":      "",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.ltiService.PrivateKeyID

	signedToken, err := token.SignedString(s.ltiService.PrivateKey)
	if err != nil {
		s.log.Error("deeplink.build_response.sign.failed", "err", err)
		return "", fmt.Errorf("failed to sign deep linking response: %w", err)
	}

	s.log.Info("deeplink.build_response.ok", "client_id", platform.ClientID, "item_count", len(items))
	return signedToken, nil
}

func (s *DeepLinkingService) CreateContentItemForLtiResource(
	title, text, url string,
	lineItem *domain.LineItem,
) *domain.ContentItem {

	s.log.Info("deeplink.create_content_item", "title", title, "has_line_item", lineItem != nil)
	item := &domain.ContentItem{
		Type:         "ltiResourceLink",
		URL:          url,
		Title:        title,
		Text:         text,
		WindowTarget: "iframe",
		IFrame: &domain.IFrame{
			Src:   url,
			Width: 800,
		},
	}

	if lineItem != nil {
		item.LineItem = lineItem
		item.Available = &domain.Availability{
			StartAt: lineItem.StartDateTime,
			EndAt:   lineItem.EndDateTime,
		}
		item.Submission = &domain.Submission{
			StartAt: lineItem.StartDateTime,
			EndAt:   lineItem.EndDateTime,
		}
	}

	return item
}

func (s *DeepLinkingService) StoreDeepLinkingSession(ctx context.Context, sessionID string, session *domain.DeepLinkingSession) error {
	s.log.Info("deeplink.session.store.start", "session_id", sessionID)
	if sessionID == "" || session == nil {
		s.log.Warn("deeplink.session.store.validation_failed")
		return fmt.Errorf("invalid session")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = session
	s.log.Info("deeplink.session.store.ok", "session_id", sessionID)
	return nil
}

func (s *DeepLinkingService) GetDeepLinkingSession(ctx context.Context, sessionID string) (*domain.DeepLinkingSession, error) {
	s.log.Info("deeplink.session.get.start", "session_id", sessionID)
	if sessionID == "" {
		s.log.Warn("deeplink.session.get.validation_failed")
		return nil, fmt.Errorf("sessionID required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.log.Warn("deeplink.session.get.not_found", "session_id", sessionID)
		return nil, fmt.Errorf("session not found")
	}
	s.log.Info("deeplink.session.get.ok", "session_id", sessionID)
	return session, nil
}

func (s *DeepLinkingService) DeleteDeepLinkingSession(ctx context.Context, sessionID string) error {
	s.log.Info("deeplink.session.delete.start", "session_id", sessionID)
	if sessionID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	s.log.Info("deeplink.session.delete.ok", "session_id", sessionID)
	return nil
}

func (s *DeepLinkingService) GetAvailableContent(ctx context.Context, contextID string, acceptTypes []string) ([]domain.ContentItem, error) {
	s.log.Info("deeplink.get_available_content.start", "context_id", contextID, "accept_types_count", len(acceptTypes))
	// TODO: wire real content source (DB/registry). For now, return a small deterministic list.
	items := []domain.ContentItem{
		{
			Type:         domain.ContentTypeLtiResourceLink,
			URL:          s.cfg.LTIConfig.Domain + "/lti/launch",
			Title:        "Demo LTI Resource",
			Text:         "Example resource for deep linking",
			WindowTarget: "iframe",
			IFrame: &domain.IFrame{
				Src:    s.cfg.LTIConfig.Domain + "/lti/launch",
				Width:  1024,
				Height: 768,
			},
			Custom: map[string]interface{}{
				"context_id": contextID,
			},
		},
		{
			Type:  domain.ContentTypeLink,
			URL:   s.cfg.LTIConfig.Domain,
			Title: "Tool Home",
			Text:  "Open tool homepage",
		},
	}

	s.log.Info("deeplink.get_available_content.ok", "context_id", contextID, "count", len(items))
	return items, nil
}

func (s *DeepLinkingService) CreateLineItem(ctx context.Context, lineItem *domain.LineItem) error {
	s.log.Info("deeplink.create_line_item.start", "context_id", func() string {
		if lineItem != nil {
			return lineItem.ContextID
		}
		return ""
	}())
	// TODO: wire to AGS service/repository if needed.
	s.log.Info("deeplink.create_line_item.ok")
	return nil
}
