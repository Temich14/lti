package service

import (
	"LTICore/internal/config"
	"LTICore/internal/core/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type DeepLinkingService struct {
	ltiService *LTIService
	cfg        *config.Config
}

func NewDeepLinkingService(ltiService *LTIService, cfg *config.Config) *DeepLinkingService {
	return &DeepLinkingService{
		ltiService: ltiService,
		cfg:        cfg,
	}
}

func (s *DeepLinkingService) HandleDeepLinkingRequest(ctx context.Context, idToken string) (*domain.DeepLinkingSettings, error) {
	token, err := s.ltiService.parseUnverifiedToken(idToken)
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)

	dlClaim, ok := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"]
	if !ok {
		return nil, fmt.Errorf("missing deep linking settings")
	}

	dlJSON, _ := json.Marshal(dlClaim)
	var settings domain.DeepLinkingSettings
	if err := json.Unmarshal(dlJSON, &settings); err != nil {
		return nil, fmt.Errorf("invalid deep linking settings: %w", err)
	}

	return &settings, nil
}

func (s *DeepLinkingService) BuildDeepLinkResponse(
	ctx context.Context,
	settings *domain.DeepLinkingSettings,
	items []domain.ContentItem,
	platform *domain.Platform,
) (string, error) {

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
		return "", fmt.Errorf("failed to sign deep linking response: %w", err)
	}

	return signedToken, nil
}

func (s *DeepLinkingService) CreateContentItemForLtiResource(
	title, text, url string,
	lineItem *domain.LineItem,
) *domain.ContentItem {

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
