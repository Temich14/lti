package service

import (
	"LTICore/internal/config"
	"LTICore/internal/core/domain"
	"context"
	"crypto/rsa"
	"time"
)

type LTIMetrics interface {
	ObserveLaunchDuration(duration time.Duration)
	IncLaunchSuccess()
	IncLaunchError()

	ObserveJWKSFetch(duration time.Duration)
	IncJWKSFetchError()

	ObserveNRPSRequest(duration time.Duration)
	IncNRPSError()
}

type LTIService struct {
	ltiClient        LTIClient
	nrpsClient       NRPSClient
	platformRepo     PlatformRepo
	loginSessionRepo LoginSessionRepository
	metrics          LTIMetrics
	cfg              *config.Config
	PrivateKey       *rsa.PrivateKey
	PrivateKeyID     string
}

func NewLtiService(client LTIClient, nrpsClient NRPSClient, platformRepo PlatformRepo, repository LoginSessionRepository, privateKey *rsa.PrivateKey, cfg *config.Config, metrics LTIMetrics) *LTIService {

	return &LTIService{ltiClient: client, nrpsClient: nrpsClient, platformRepo: platformRepo, loginSessionRepo: repository, PrivateKey: privateKey, PrivateKeyID: cfg.OAuthConfig.KeyID, cfg: cfg, metrics: metrics}
}

type AGSRepository interface {
	GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error)
	CreateLineItem(ctx context.Context, li *domain.LineItem) error
	GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error)
	SaveScore(ctx context.Context, score *domain.Score) error
}

type LoginSessionRepository interface {
	Save(ctx context.Context, session *domain.LoginSession) error
	GetAndDelete(ctx context.Context, state string) (*domain.LoginSession, error)
}

type PlatformRepo interface {
	GetByIssuerAndClientID(ctx context.Context, issuer, clientID string) (*domain.Platform, error)
	Create(ctx context.Context, platform *domain.Platform) (*domain.Platform, error)
}

type LTIClient interface {
	GetOpenidConfiguration(openidUrl string) (*domain.OpenidConfiguration, error)
	SendRegistrationRequest(registrationUrl, registrationToken string, body *domain.ToolRegistrationRequest) (*domain.ToolRegistrationResponse, error)
	GetAccessToken(ctx context.Context, platform *domain.Platform, scope, privateKeyID string, key *rsa.PrivateKey) (string, error)
}

type NRPSClient interface {
	GetMembers(ctx context.Context, url string, token string) ([]domain.NRPSMember, error)
}
