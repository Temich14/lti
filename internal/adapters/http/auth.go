package http

import (
	"LTICore/internal/core/domain"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PlatformCfg struct {
	Issuer                string
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

type LTIService interface {
	Register(ctx context.Context, openidUrl, registrationToken string) error
	Login(ctx context.Context, req *domain.LoginRequest) (string, error)
	Launch(ctx context.Context, idToken, state string) ([]domain.NRPSMember, error)
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

func (a *AuthAdapter) Login(c *gin.Context) {
	var req domain.LoginRequest
	err := c.ShouldBind(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	loginUrl, err := a.ltiService.Login(c.Request.Context(), &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.Redirect(302, loginUrl)
}

func (a *AuthAdapter) Register(c *gin.Context) {
	openidUrl := c.Query("openid_configuration")
	registrationToken := c.Query("registration_token")

	err := a.ltiService.Register(c, openidUrl, registrationToken)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	c.JSON(200, gin.H{})
}
