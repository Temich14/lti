package http

import (
	"LTICore/internal/core/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type PlatformCfg struct {
	Issuer                string
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

type LTIService interface {
	Register(ctx context.Context, openidUrl, registrationToken string) error
	Login(ctx context.Context, req *domain.LoginRequest) (string, error)
	Launch(ctx context.Context, idToken, state string) (*domain.LaunchContext, error)
}

type AuthAdapter struct {
	ltiService LTIService
}

func NewAuthAdapter(ltiService LTIService) *AuthAdapter {
	return &AuthAdapter{ltiService: ltiService}
}

type RegistrationRequest struct {
	OpenIDConfig      string `json:"openid_configuration" form:"openid_configuration"`
	RegistrationToken string `json:"registration_token" form:"registration_token"`
}

// OIDCLogin implements LTI 1.3 OIDC login initiation endpoint.
// It accepts both GET (query params) and POST (form params).
func (a *AuthAdapter) OIDCLogin(c *gin.Context) {
	iss := c.Query("iss")
	if iss == "" {
		iss = c.PostForm("iss")
	}

	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = c.PostForm("client_id")
	}

	loginHintRaw := c.Query("login_hint")
	if loginHintRaw == "" {
		loginHintRaw = c.PostForm("login_hint")
	}

	ltiMessageHintRaw := c.Query("lti_message_hint")
	if ltiMessageHintRaw == "" {
		ltiMessageHintRaw = c.PostForm("lti_message_hint")
	}

	// Spec name is target_link_uri. Our domain model uses a legacy field/tag.
	targetLinkURI := c.Query("target_link_uri")
	if targetLinkURI == "" {
		targetLinkURI = c.PostForm("target_link_uri")
	}
	if targetLinkURI == "" {
		targetLinkURI = c.Query("target_ling_uri")
		if targetLinkURI == "" {
			targetLinkURI = c.PostForm("target_ling_uri")
		}
	}

	if iss == "" || clientID == "" || loginHintRaw == "" || targetLinkURI == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "iss, client_id, login_hint, target_link_uri are required"})
		return
	}

	loginHint, _ := strconv.Atoi(loginHintRaw)
	ltiMessageHint, _ := strconv.Atoi(ltiMessageHintRaw)

	req := domain.LoginRequest{
		Iss:            iss,
		ClientID:       clientID,
		LoginHint:      loginHint,
		LTIMessageHint: ltiMessageHint,
		TargetLinkURI:  targetLinkURI,
		TargetLingUri:  targetLinkURI,
	}

	loginURL, err := a.ltiService.Login(c.Request.Context(), &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusSeeOther, loginURL)
}

func (a *AuthAdapter) Login(c *gin.Context) {
	var req domain.LoginRequest
	err := c.ShouldBind(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginUrl, err := a.ltiService.Login(c.Request.Context(), &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(302, loginUrl)
}

func (a *AuthAdapter) Register(c *gin.Context) {
	openidUrl := c.Query("openid_configuration")
	registrationToken := c.Query("registration_token")
	if openidUrl == "" || registrationToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "openid_configuration and registration_token are required"})
		return
	}

	err := a.ltiService.Register(c.Request.Context(), openidUrl, registrationToken)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html>
<head><title>Registration Complete</title></head>
<body>
<script>
  (window.opener || window.parent).postMessage(
    { subject: 'org.imsglobal.lti.close' },
    '*'
  );
</script>
<p>Registration completed. You can close this window.</p>
</body>
</html>
`))
}
