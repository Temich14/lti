package app

import (
	"LTICore/internal/adapters/http"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Server struct {
	api           *gin.Engine
	authHandler   *http.AuthAdapter
	launchHandler *http.LaunchAdapter
	jwksHandler   *http.JWKSHandler
	agsHandler    *http.AGSHandler
}

func NewServer(api *gin.Engine, authHandler *http.AuthAdapter, launchHandler *http.LaunchAdapter, jwkHandler *http.JWKSHandler, agsHandler *http.AGSHandler) *Server {
	return &Server{api: api, authHandler: authHandler, launchHandler: launchHandler, jwksHandler: jwkHandler, agsHandler: agsHandler}
}

func (s *Server) RegisterRoutes() {
	group := s.api.Group("/lti")

	dlGroup := group.Group("/deeplink")
	dlGroup.Get("/lti/deeplink/select", dlHandler.ShowContentSelection)
    dlGroup.Post("/lti/deeplink/return", dlHandler.ReturnContent)
	
	agsGroup := group.Group("/ags")
	agsGroup.GET("/lineitems", s.agsHandler.GetScore)
	agsGroup.POST("/lineitems", s.agsHandler.PostScore)
	agsGroup.GET("/score", s.agsHandler.GetScore)
	agsGroup.POST("/score", s.agsHandler.PostScore)

	group.GET("/.well-known/jwks.json", s.jwksHandler.JWKS)

	launchGroup := group.Group("/launch")
	launchGroup.POST("", s.launchHandler.Launch)

	authGroup := group.Group("/auth")
	authGroup.GET("/register", s.authHandler.Register)
	authGroup.POST("/login", s.authHandler.Login)

	for _, route := range s.api.Routes() {
		fmt.Printf("Route: %s %s\n", route.Method, route.Path)
	}
}
