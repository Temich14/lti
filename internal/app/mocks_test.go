package app

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rsa"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockAGSRepository struct {
	mock.Mock
}

func (m *MockAGSRepository) GetLineItems(ctx context.Context, contextID string) ([]domain.LineItem, error) {
	args := m.Called(ctx, contextID)
	return args.Get(0).([]domain.LineItem), args.Error(1)
}

func (m *MockAGSRepository) CreateLineItem(ctx context.Context, li *domain.LineItem) error {
	args := m.Called(ctx, li)
	return args.Error(0)
}

func (m *MockAGSRepository) GetScore(ctx context.Context, lineItemID, userID string) (*domain.Score, error) {
	args := m.Called(ctx, lineItemID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Score), args.Error(1)
}

func (m *MockAGSRepository) SaveScore(ctx context.Context, score *domain.Score) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

type MockAGSMetrics struct {
	mock.Mock
}

func (m *MockAGSMetrics) IncAGSRequestErrors() { m.Called() }
func (m *MockAGSMetrics) IncAGSRequestsTotal() { m.Called() }

type MockLTIClient struct {
	mock.Mock
}

func (m *MockLTIClient) GetOpenidConfiguration(openidUrl string) (*domain.OpenidConfiguration, error) {
	args := m.Called(openidUrl)
	return args.Get(0).(*domain.OpenidConfiguration), args.Error(1)
}

func (m *MockLTIClient) SendRegistrationRequest(registrationUrl, registrationToken string, body *domain.ToolRegistrationRequest) (*domain.ToolRegistrationResponse, error) {
	args := m.Called(registrationUrl, registrationToken, body)
	return args.Get(0).(*domain.ToolRegistrationResponse), args.Error(1)
}

func (m *MockLTIClient) GetAccessToken(ctx context.Context, platform *domain.Platform, scope, privateKeyID string, key *rsa.PrivateKey) (string, error) {
	args := m.Called(ctx, platform, scope, privateKeyID, key)
	return args.String(0), args.Error(1)
}

type MockNRPSClient struct {
	mock.Mock
}

func (m *MockNRPSClient) GetMembers(ctx context.Context, url string, token string) ([]domain.NRPSMember, error) {
	args := m.Called(ctx, url, token)
	return args.Get(0).([]domain.NRPSMember), args.Error(1)
}

func (m *MockNRPSClient) GetMembersPage(ctx context.Context, url string, token string) ([]domain.NRPSMember, string, error) {
	args := m.Called(ctx, url, token)
	return args.Get(0).([]domain.NRPSMember), args.Get(1).(string), args.Error(2)
}

type MockPlatformRepo struct {
	mock.Mock
}

func (m *MockPlatformRepo) GetByIssuerAndClientID(ctx context.Context, issuer, clientID string) (*domain.Platform, error) {
	args := m.Called(ctx, issuer, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Platform), args.Error(1)
}

func (m *MockPlatformRepo) Create(ctx context.Context, platform *domain.Platform) (*domain.Platform, error) {
	args := m.Called(ctx, platform)
	return args.Get(0).(*domain.Platform), args.Error(1)
}

type MockLoginSessionRepo struct {
	mock.Mock
}

func (m *MockLoginSessionRepo) Save(ctx context.Context, session *domain.LoginSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockLoginSessionRepo) GetAndDelete(ctx context.Context, state string) (*domain.LoginSession, error) {
	args := m.Called(ctx, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoginSession), args.Error(1)
}

type MockLTIMetrics struct {
	mock.Mock
}

func (m *MockLTIMetrics) ObserveLaunchDuration(duration time.Duration) { m.Called(duration) }
func (m *MockLTIMetrics) IncLaunchSuccess()                            { m.Called() }
func (m *MockLTIMetrics) IncLaunchError()                              { m.Called() }
func (m *MockLTIMetrics) ObserveJWKSFetch(duration time.Duration)      { m.Called(duration) }
func (m *MockLTIMetrics) IncJWKSFetchError()                           { m.Called() }
func (m *MockLTIMetrics) ObserveNRPSRequest(duration time.Duration)    { m.Called(duration) }
func (m *MockLTIMetrics) IncNRPSError()                                { m.Called() }

type MockDeepLinkingService struct {
	mock.Mock
}

func (m *MockDeepLinkingService) HandleDeepLinkingRequest(ctx context.Context, idToken string) (*domain.DeepLinkingSettings, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeepLinkingSettings), args.Error(1)
}

func (m *MockDeepLinkingService) BuildDeepLinkResponse(ctx context.Context, settings *domain.DeepLinkingSettings, items []domain.ContentItem, platform *domain.Platform) (string, error) {
	args := m.Called(ctx, settings, items, platform)
	return args.String(0), args.Error(1)
}

func (m *MockDeepLinkingService) CreateContentItemForLtiResource(title, text, url string, lineItem *domain.LineItem) *domain.ContentItem {
	args := m.Called(title, text, url, lineItem)
	return args.Get(0).(*domain.ContentItem)
}

func (m *MockDeepLinkingService) StoreDeepLinkingSession(ctx context.Context, sessionID string, session *domain.DeepLinkingSession) error {
	args := m.Called(ctx, sessionID, session)
	return args.Error(0)
}

func (m *MockDeepLinkingService) GetDeepLinkingSession(ctx context.Context, sessionID string) (*domain.DeepLinkingSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeepLinkingSession), args.Error(1)
}

func (m *MockDeepLinkingService) DeleteDeepLinkingSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockDeepLinkingService) GetAvailableContent(ctx context.Context, contextID string, acceptTypes []string) ([]domain.ContentItem, error) {
	args := m.Called(ctx, contextID, acceptTypes)
	return args.Get(0).([]domain.ContentItem), args.Error(1)
}

func (m *MockDeepLinkingService) CreateLineItem(ctx context.Context, lineItem *domain.LineItem) error {
	args := m.Called(ctx, lineItem)
	return args.Error(0)
}
