package http

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type LaunchAdapter struct {
	ltiService LTIService
	dlService  LaunchDeepLinkingService
	enrollStarter EnrollmentSyncStarter
}

type LaunchDeepLinkingService interface {
	StoreDeepLinkingSession(ctx context.Context, sessionID string, session *domain.DeepLinkingSession) error
}

type EnrollmentSyncStarter interface {
	StartOrResumeRosterSync(ctx context.Context, issuer, clientID, lmsCourseID, nrpsContextMembershipsURL string) (uuid.UUID, error)
}

func NewLaunchAdapter(ltiService LTIService, dlService LaunchDeepLinkingService, enrollStarter EnrollmentSyncStarter) *LaunchAdapter {
	return &LaunchAdapter{ltiService: ltiService, dlService: dlService, enrollStarter: enrollStarter}
}

type LaunchRequest struct {
	IdToken string `json:"id_token"`
	State   string `json:"state"`
}

func (l *LaunchAdapter) Launch(c *gin.Context) {

	idToken := c.PostForm("id_token")
	state := c.PostForm("state")

	ctxData, err := l.ltiService.Launch(c.Request.Context(), idToken, state)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// extract DL settings from claims (IMPORTANT FIX)
	dlSettings := extractDeepLinkingSettings(ctxData.Claims)

	// Start enrollment synchronization in background (NRPS roster walk -> outbox -> Kafka).
	if l.enrollStarter != nil {
		if courseID := getContextID(ctxData.Claims); courseID != "" {
			if nrpsClaimRaw, ok := ctxData.Claims["https://purl.imsglobal.org/spec/lti-nrps/claim/namesroleservice"]; ok {
				if nrpsJSON, err := json.Marshal(nrpsClaimRaw); err == nil {
					var nrps domain.NRPSClaim
					if err := json.Unmarshal(nrpsJSON, &nrps); err == nil && nrps.ContextMembershipsURL != "" {
						_, _ = l.enrollStarter.StartOrResumeRosterSync(
							c.Request.Context(),
							ctxData.Platform.Issuer,
							ctxData.Platform.ClientID,
							courseID,
							nrps.ContextMembershipsURL,
						)
					}
				}
			}
		}
	}

	sessionID := generateLaunchSessionID()

	_ = l.dlService.StoreDeepLinkingSession(
		c.Request.Context(),
		sessionID,
		&domain.DeepLinkingSession{
			Settings:  dlSettings,
			Platform:  ctxData.Platform,
			UserID:    getUserID(ctxData.Claims),
			ContextID: getContextID(ctxData.Claims),
		},
	)

	selectRedirectURL := fmt.Sprintf("deeplink/select?session_id=%s", sessionID)
	c.Redirect(http.StatusSeeOther, selectRedirectURL)
}
func getUserID(claims jwt.MapClaims) string {
	if claims == nil {
		return ""
	}

	if sub, ok := claims["sub"].(string); ok {
		return sub
	}

	return ""
}
func getContextID(claims jwt.MapClaims) string {
	if claims == nil {
		return ""
	}

	raw, ok := claims["https://purl.imsglobal.org/spec/lti/claim/context"]
	if !ok {
		return ""
	}

	ctx, ok := raw.(map[string]interface{})
	if !ok {
		return ""
	}

	if id, ok := ctx["id"].(string); ok {
		return id
	}

	return ""
}
func extractDeepLinkingSettings(claims jwt.MapClaims) *domain.DeepLinkingSettings {

	raw, ok := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"]
	if !ok {
		return &domain.DeepLinkingSettings{}
	}

	settingsMap, ok := raw.(map[string]interface{})
	if !ok {
		return &domain.DeepLinkingSettings{}
	}

	settings := &domain.DeepLinkingSettings{}

	if v, ok := settingsMap["accept_types"].([]interface{}); ok {
		settings.AcceptTypes = toStringSlice(v)
	}

	if v, ok := settingsMap["accept_presentation_document_targets"].([]interface{}); ok {
		settings.AcceptPresentationDocumentTargets = toStringSlice(v)
	}

	if v, ok := settingsMap["accept_multiple"].(bool); ok {
		settings.AcceptMultiple = v
	}

	if v, ok := settingsMap["auto_create"].(bool); ok {
		settings.AutoCreate = v
	}

	if v, ok := settingsMap["title"].(string); ok {
		settings.Title = v
	}

	if v, ok := settingsMap["text"].(string); ok {
		settings.Text = v
	}

	if v, ok := settingsMap["deep_link_return_url"].(string); ok {
		settings.DeepLinkReturnURL = v
	}

	return settings
}

func toStringSlice(input []interface{}) []string {
	out := make([]string, 0, len(input))
	for _, v := range input {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
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
