package service

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type deepLinkSessionStore interface {
	StoreDeepLinkingSession(ctx context.Context, sessionID string, session *domain.DeepLinkingSession) error
}

// LaunchFlowService runs everything that must happen once an LTI id_token validates (beyond pure OIDC semantics).
type LaunchFlowService struct {
	log    *slog.Logger
	lti    *LTIService
	dl     deepLinkSessionStore
	enroll EnrollmentSyncStarter
	edu    EducationSessionCreator
}

func NewLaunchFlowService(
	logger *slog.Logger,
	lti *LTIService,
	dl deepLinkSessionStore,
	enroll EnrollmentSyncStarter,
	edu EducationSessionCreator,
) *LaunchFlowService {
	if logger == nil {
		logger = slog.Default()
	}
	return &LaunchFlowService{log: logger, lti: lti, dl: dl, enroll: enroll, edu: edu}
}

// Run executes LTI OIDC verification, sync kick-off, session registration, deep-link stash, and returns the deeplink-relative redirect URL.
func (s *LaunchFlowService) Run(ctx context.Context, idToken, state string) (deeplinkRelativeURL string, err error) {
	if s.lti == nil || s.dl == nil {
		return "", fmt.Errorf("launch flow is not fully configured")
	}
	ctxData, err := s.lti.Launch(ctx, idToken, state)
	if err != nil {
		return "", err
	}

	issuer := ""
	clientID := ""
	if ctxData.Platform != nil {
		issuer = ctxData.Platform.Issuer
		clientID = ctxData.Platform.ClientID
	}

	dlSettings := domain.ClaimsDeepLinkingSettings(ctxData.Claims)
	userID := domain.ClaimsUserSub(ctxData.Claims)
	contextID := domain.ClaimsContextID(ctxData.Claims)

	s.startEnrollmentWithoutCancel(ctx, issuer, clientID, contextID, ctxData)

	if s.edu != nil && userID != "" && contextID != "" {
		reqID := uuid.NewString()
		p := EducationSessionCreateParams{
			RequestID:  reqID,
			UserID:     userID,
			CourseID:   contextID,
			ResourceID: domain.ClaimsResourceLinkID(ctxData.Claims),
		}
		if sid, err := s.edu.CreateForLTILaunch(ctx, p); err != nil {
			s.log.Warn("education.session.create.failed", "err", err)
		} else if sid != "" {
			s.log.Info("education.session.created", "session_id", sid, "request_id", reqID)
		}
	}

	sessionID := newLaunchBrowseSessionID()
	if err := s.dl.StoreDeepLinkingSession(
		ctx,
		sessionID,
		&domain.DeepLinkingSession{
			Settings:  dlSettings,
			Platform:  ctxData.Platform,
			UserID:    userID,
			ContextID: contextID,
		},
	); err != nil {
		s.log.Warn("deeplink.store_session.failed", "err", err)
	}

	return fmt.Sprintf("deeplink/select?session_id=%s", sessionID), nil
}

func (s *LaunchFlowService) startEnrollmentWithoutCancel(
	ctx context.Context,
	issuer string,
	clientID string,
	contextID string,
	ctxData *domain.LaunchContext,
) {
	if s.enroll == nil || ctxData == nil || contextID == "" || issuer == "" || clientID == "" {
		return
	}
	nrpsURL := domain.ClaimsNRPSMembershipsURL(ctxData.Claims)
	if nrpsURL == "" {
		return
	}
	bg := detachedCtx(ctx)
	_, syncErr := s.enroll.StartOrResumeRosterSync(bg, issuer, clientID, contextID, nrpsURL)
	if syncErr != nil {
		s.log.Warn("enrollment.roster_sync.start.failed", "err", syncErr)
	}
}

func detachedCtx(parent context.Context) context.Context {
	// Enrollment should not be tied to POST body stream lifetime once launch logic has committed.
	if parent == nil {
		return context.Background()
	}
	return context.WithoutCancel(parent)
}

func newLaunchBrowseSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
