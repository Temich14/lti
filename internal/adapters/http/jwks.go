package http

import (
	"LTICore/internal/core/domain"
	"github.com/gin-gonic/gin"
)

type JwksService interface {
	PrepareJwks() *domain.JWKS
}
type JWKSHandler struct {
	service JwksService
}

func NewJWKSHandler(service JwksService) *JWKSHandler {
	return &JWKSHandler{service: service}
}

func (h *JWKSHandler) JWKS(c *gin.Context) {
	jwks := h.service.PrepareJwks()
	c.JSON(200, jwks)
}
