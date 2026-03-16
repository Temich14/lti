package http

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

type LaunchAdapter struct {
	ltiService LTIService
	dlService  LaunchDeepLinkingService
}

type LaunchDeepLinkingService interface {
	StoreDeepLinkingSession(ctx context.Context, sessionID string, session *domain.DeepLinkingSession) error
}

func NewLaunchAdapter(ltiService LTIService, dlService LaunchDeepLinkingService) *LaunchAdapter {
	return &LaunchAdapter{ltiService: ltiService, dlService: dlService}
}

type LaunchRequest struct {
	IdToken string `json:"id_token"`
	State   string `json:"state"`
}

func (l *LaunchAdapter) Launch(c *gin.Context) {
	var req LaunchRequest
	req.IdToken = c.PostForm("id_token")
	req.State = c.PostForm("state")

	members, err := l.ltiService.Launch(c.Request.Context(), req.IdToken, req.State)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = members

	// Create deeplink session for selector UI.
	sessionID := generateLaunchSessionID()
	settings := &domain.DeepLinkingSettings{
		AcceptTypes:       []string{domain.ContentTypeLtiResourceLink},
		AcceptMultiple:    true,
		Title:             "Select Content",
		Text:              "Choose an item to deeplink.",
		DeepLinkReturnURL: c.Query("return_url"),
	}

	platform := extractPlatformFromIDToken(req.IdToken)
	if platform == nil {
		platform = &domain.Platform{}
	}

	_ = l.dlService.StoreDeepLinkingSession(c.Request.Context(), sessionID, &domain.DeepLinkingSession{
		Settings:  settings,
		Platform:  platform,
		UserID:    "",
		ContextID: "",
	})

	c.SetCookie("dl_session", sessionID, 3600, "/", "", false, true)

	c.HTML(http.StatusOK, "deeplinking/select.html", gin.H{
		"acceptMultiple": settings.AcceptMultiple,
		"acceptTypes":    settings.AcceptTypes,
		"title":          settings.Title,
		"text":           settings.Text,
		"returnUrl":      settings.DeepLinkReturnURL,
	})
}

func generateLaunchSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func extractPlatformFromIDToken(idToken string) *domain.Platform {
	parser := jwt.Parser{}
	token, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	issuer, _ := claims["iss"].(string)
	// aud can be string or array
	clientID, _ := claims["aud"].(string)
	if clientID == "" {
		if arr, ok := claims["aud"].([]interface{}); ok && len(arr) > 0 {
			if s, ok := arr[0].(string); ok {
				clientID = s
			}
		}
	}

	var ctxID string
	if raw, ok := claims["https://purl.imsglobal.org/spec/lti/claim/context"]; ok {
		if b, err := json.Marshal(raw); err == nil {
			var tmp struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(b, &tmp); err == nil {
				ctxID = tmp.ID
			}
		}
	}

	p := &domain.Platform{Issuer: issuer, ClientID: clientID}
	_ = ctxID
	return p
}
