package app

import (
	"LTICore/internal/adapters/http"
	"github.com/gin-gonic/gin"
)

type Server struct {
	api           *gin.Engine
	authHandler   *http.AuthAdapter
	launchHandler *http.LaunchAdapter
	jwksHandler   *http.JWKSHandler
	agsHandler    *http.AGSHandler
	dlHandler     *http.DeepLinkingHandler
}

func NewServer(api *gin.Engine, authHandler *http.AuthAdapter, launchHandler *http.LaunchAdapter, jwkHandler *http.JWKSHandler, agsHandler *http.AGSHandler, dlHandler *http.DeepLinkingHandler) *Server {
	return &Server{
		api:           api,
		authHandler:   authHandler,
		launchHandler: launchHandler,
		jwksHandler:   jwkHandler,
		agsHandler:    agsHandler,
		dlHandler:     dlHandler,
	}
}

func (s *Server) RegisterRoutes() {
	group := s.api.Group("/lti")

	dlGroup := group.Group("/deeplink")
	dlGroup.POST("/select", s.dlHandler.ShowContentSelection)
	dlGroup.POST("/return", s.dlHandler.ReturnContent)
	dlGroup.POST("/api/return", s.dlHandler.APIReturnContent)
	dlGroup.GET("/content", s.dlHandler.GetAvailableContent)
	dlGroup.POST("/lineitem", s.dlHandler.CreateLineItemFromSelection)
	dlGroup.GET("/cancel", s.dlHandler.CancelDeepLinking)

	agsGroup := group.Group("/ags")
	agsGroup.GET("/lineitems", s.agsHandler.GetLineItems)
	agsGroup.POST("/lineitems", s.agsHandler.CreateLineItem)
	agsGroup.GET("/score", s.agsHandler.GetScore)
	agsGroup.POST("/score", s.agsHandler.PostScore)

	group.GET("/.well-known/jwks.json", s.jwksHandler.JWKS)

	launchGroup := group.Group("/launch")
	launchGroup.POST("", s.launchHandler.Launch)

	authGroup := group.Group("/auth")
	authGroup.GET("/register", s.authHandler.Register)
	authGroup.POST("/login", s.authHandler.Login)

	_ = s.api.Routes()
}
