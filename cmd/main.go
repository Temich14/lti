package main

import (
	"LTICore/internal/adapters/http"
	"LTICore/internal/app"
	"LTICore/internal/config"
	"LTICore/internal/core/service"
	"LTICore/internal/infrastructure/db"
	http2 "LTICore/internal/infrastructure/http"
	"LTICore/internal/infrastructure/metrics"
	"LTICore/internal/infrastructure/repo"
	"context"
	"encoding/json"
	"html/template"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample rk-demo server.
// @termsOfService http://swagger.io/terms/

// @securityDefinitions.basic BasicAuth

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {

	mtrcs := metrics.NewLTIMetrics()

	boot := rkboot.NewBoot()
	ctx := context.Background()

	boot.Bootstrap(ctx)

	cfg, err := config.Load("boot.yaml")
	if err != nil {
		panic(err)
	}
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	pool, err := db.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	jwkPK, err := config.LoadPrivateKey(cfg.OAuthConfig.PrivateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	templatesPath := "templates"
	ginEntry := rkgin.GetGinEntry("lti-core")
	ginEngine := ginEntry.Router
	ginEngine.SetFuncMap(template.FuncMap{
		"toJson": func(v interface{}) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "null"
			}
			return string(b)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	})
	templatePattern := filepath.Join(templatesPath, "*", "*.html")
	log.Printf("Loading templates from pattern: %s", templatePattern)
	ginEngine.LoadHTMLGlob("templates/deeplinking/*.html")

	ginEntry.Router.Static("/static", "./static")
	err = mtrcs.Register(ginEntry.PromEntry.Registerer)
	if err != nil {
		log.Fatal("failed to register custom metrics:", err)
	}
	queries := db.New(pool)

	loginSessionRepo := repo.NewRedisClient(&cfg.RedisConfig)
	platformRepo := repo.NewPlatformRepository(queries)
	nrpsClient := http2.NewNRPSClient()
	ltiClient := http2.NewLtiClient(logger)

	jwkservice := service.NewJwksService(cfg, jwkPK)
	agsRepo := repo.NewMockAgsRepo()
	agsService := service.NewAGSService(agsRepo, mtrcs)
	ltiService := service.NewLtiService(ltiClient, nrpsClient, platformRepo, loginSessionRepo, jwkPK, cfg, mtrcs)
	dlService := service.NewDeepLinkingService(ltiService, cfg)

	authAdapter := http.NewAuthAdapter(ltiService)
	launchAdapter := http.NewLaunchAdapter(ltiService, dlService)
	jwkHandler := http.NewJWKSHandler(jwkservice)
	agsHandler := http.NewAGSHandler(agsService)
	dlHandler := http.NewDeepLinkingHandler(dlService)

	server := app.NewServer(ginEntry.Router, authAdapter, launchAdapter, jwkHandler, agsHandler, dlHandler)
	server.RegisterRoutes()

	boot.WaitForShutdownSig(ctx)

	pool.Close()
}

func LoadAppConfigRaw(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return raw, nil
}
