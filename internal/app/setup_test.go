package app

import (
	adaptershttp "LTICore/internal/adapters/http"
	"LTICore/internal/core/domain"
	"LTICore/internal/core/service"
	"encoding/json"
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestServer(t *testing.T) (
	*Server,
	*MockAGSRepository,
	*MockAGSMetrics,
	*MockLTIClient,
	*MockNRPSClient,
	*MockPlatformRepo,
	*MockLoginSessionRepo,
	*MockLTIMetrics,
	*MockDeepLinkingService,
) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	agsRepo := new(MockAGSRepository)
	agsMetrics := new(MockAGSMetrics)
	agsService := service.NewAGSService(agsRepo, agsMetrics)
	agsHandler := adaptershttp.NewAGSHandler(agsService)

	ltiClient := new(MockLTIClient)
	nrpsClient := new(MockNRPSClient)
	platformRepo := new(MockPlatformRepo)
	loginSessionRepo := new(MockLoginSessionRepo)
	ltiMetrics := new(MockLTIMetrics)

	ltiService := service.NewLtiService(
		ltiClient, nrpsClient, platformRepo, loginSessionRepo,
		testPrivateKey, testConfig, ltiMetrics,
	)

	dlService := new(MockDeepLinkingService)
	dlHandler := adaptershttp.NewDeepLinkingHandler(dlService)

	authHandler := adaptershttp.NewAuthAdapter(ltiService)
	launchHandler := adaptershttp.NewLaunchAdapter(ltiService, dlService)
	jwksHandler := adaptershttp.NewJWKSHandler(service.NewJwksService(testConfig, testPrivateKey))

	engine := gin.New()
	tmpl := template.New("").Funcs(template.FuncMap{
		"json": func(v any) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	})
	template.Must(tmpl.New("error.html").Parse("{{.error}}"))
	template.Must(tmpl.New("deeplink/select.html").Parse("{{.title}}"))
	template.Must(tmpl.New("deeplink/return.html").Parse("{{.jwt}}{{.error}}{{.returnUrl}}"))
	engine.SetHTMLTemplate(tmpl)
	engine.Use(func(c *gin.Context) {
		// minimal context required by deep linking handlers
		c.Set("platform", &domain.Platform{ClientID: "client123"})
		c.Set("user_id", "user-1")
		c.Set("context_id", "context-123")
		c.Next()
	})
	server := NewServer(engine, authHandler, launchHandler, jwksHandler, agsHandler, dlHandler)
	server.RegisterRoutes()

	// sanity: ensure router exists
	_ = httptest.NewRecorder()

	return server, agsRepo, agsMetrics, ltiClient, nrpsClient, platformRepo, loginSessionRepo, ltiMetrics, dlService
}
