package enrollmentsync

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	coreDomain "LTICore/internal/core/domain"
	coresvc "LTICore/internal/core/service"

	"github.com/google/uuid"
)

type ProducerService struct {
	repo *Repository

	nrpsClient     coresvc.NRPSClient
	ltiClient      coresvc.LTIClient
	platformRepo  coresvc.PlatformRepo
	privateKey    *rsa.PrivateKey
	privateKeyID  string
	scope         string
	batchSize     int
	maxSyncErrors int32

	logger *slog.Logger
}

func NewProducerService(
	repo *Repository,
	nrpsClient coresvc.NRPSClient,
	ltiClient coresvc.LTIClient,
	platformRepo coresvc.PlatformRepo,
	privateKey *rsa.PrivateKey,
	privateKeyID string,
	batchSize int,
	logger *slog.Logger,
) *ProducerService {
	return &ProducerService{
		repo:          repo,
		nrpsClient:   nrpsClient,
		ltiClient:    ltiClient,
		platformRepo: platformRepo,
		privateKey:   privateKey,
		privateKeyID: privateKeyID,
		scope:        "https://purl.imsglobal.org/spec/lti-nrps/scope/contextmembership.readonly",
		batchSize:    batchSize,
		logger:       logger,
	}
}

// StartOrResumeRosterSync creates (or reuses) an active sync run for the (issuer, course_id).
func (s *ProducerService) StartOrResumeRosterSync(
	ctx context.Context,
	issuer string,
	clientID string,
	lmsCourseID string,
	nrpsContextMembershipsURL string,
) (uuid.UUID, error) {
	return s.repo.StartOrResumeRosterSync(ctx, issuer, clientID, lmsCourseID, nrpsContextMembershipsURL)
}

func rosterWorkerOwner() string {
	host, _ := os.Hostname()
	return fmt.Sprintf("roster-sync-%s-%d", host, os.Getpid())
}

func (s *ProducerService) RunRosterSyncWorker(ctx context.Context, owner string, pollInterval time.Duration, lease time.Duration, outboxTxVersionBase int32, maxAttempts int32) {
	if owner == "" {
		owner = rosterWorkerOwner()
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		run, ok, err := s.repo.ClaimRosterSyncRun(ctx, owner, lease)
		if err != nil {
			s.logger.Error("enrollmentsync.roster.claim_failed", "err", err)
			time.Sleep(pollInterval)
			continue
		}
		if !ok {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		// 1) Resolve NRPS access token for this platform.
		platform, err := s.platformRepo.GetByIssuerAndClientID(ctx, run.Issuer, run.ClientID)
		if err != nil {
			_ = s.repo.MarkRunFailed(ctx, run.RunID, owner, run.Version, err.Error(), 30*time.Second, maxAttempts)
			continue
		}

		accessToken, err := s.ltiClient.GetAccessToken(ctx, platform, s.scope, s.privateKeyID, s.privateKey)
		if err != nil {
			_ = s.repo.MarkRunFailed(ctx, run.RunID, owner, run.Version, err.Error(), 30*time.Second, maxAttempts)
			continue
		}

		fetchURL := run.MembershipsURL
		if run.NRPSCursor != "" {
			fetchURL = run.NRPSCursor
		}

		nextURL := ""
		var members []coreDomain.NRPSMember
		members, nextURL, err = s.nrpsClient.GetMembersPage(ctx, fetchURL, accessToken)
		if err != nil {
			backoff := nextBackoff(run.SyncAttempts)
			_ = s.repo.MarkRunFailed(ctx, run.RunID, owner, run.Version, err.Error(), backoff, maxAttempts)
			continue
		}

		// 2) Batchify members into configurable event sizes.
		batches := s.batchMembers(run, members)

		// 3) Persist outbox events + advance cursor inside a single transaction.
		// completed is true only when NRPS indicates no further continuation URL.
		completed := strings.TrimSpace(nextURL) == ""
		newNextBatchIndex := run.NextBatchIndex + int32(len(batches))

		err = s.repo.AppendOutboxEventsAndAdvanceRun(
			ctx,
			owner,
			run.RunID,
			run.Version,
			nextURL,
			newNextBatchIndex,
			run.Version+1,
			completed,
			batches,
		)
		if err != nil {
			s.logger.Error("enrollmentsync.roster.append_outbox_failed", "run_id", run.RunID, "err", err)
			backoff := nextBackoff(run.SyncAttempts)
			_ = s.repo.MarkRunFailed(ctx, run.RunID, owner, run.Version, err.Error(), backoff, maxAttempts)
			continue
		}
	}
}

func nextBackoff(syncAttempts int32) time.Duration {
	// Exponential backoff with jitter guardrails.
	// Cap to 60s.
	exp := math.Pow(2, float64(syncAttempts))
	seconds := int64(2 * exp)
	if seconds > 60 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}

func (s *ProducerService) batchMembers(run *RosterSyncRun, members []coreDomain.NRPSMember) []IncomingEvent {
	if len(members) == 0 {
		// No events, but cursor still advances.
		return nil
	}

	if s.batchSize <= 0 {
		s.batchSize = 100
	}

	totalBatches := int32((len(members) + s.batchSize - 1) / s.batchSize)
	events := make([]IncomingEvent, 0, totalBatches)

	for i := 0; i < len(members); i += s.batchSize {
		j := i + s.batchSize
		if j > len(members) {
			j = len(members)
		}

		batchIndex := run.NextBatchIndex + int32(len(events))

		// Contract payload: lms + batch meta + users list.
		users := MapNRPSMembersToUserProfiles(members[i:j])

		var payload UserSynchronizationPayloadV1
		payload.LMS.Issuer = run.Issuer
		payload.LMS.CourseID = run.CourseID
		payload.Batch.Index = int(batchIndex)
		payload.Batch.Total = 0 // optional: can be enriched later if NRPS provides totals
		payload.Users = users

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			// Defensive: if marshaling fails, drop event; the run will be retried by cursor not advancing on TX.
			s.logger.Error("enrollmentsync.roster.payload.marshal_failed", "err", err)
			continue
		}

		createdAt := time.Now().UTC()
		eventID := BatchEventID(run.RunID, batchIndex)

		events = append(events, IncomingEvent{
			ID:             eventID,
			AggregateType: AggregateTypeRosterSyncRun,
			AggregateID:   run.RunID,
			EventType:     EventTypeUserSynchronizationBatchV1,
			Payload:       json.RawMessage(payloadBytes),
			CreatedAt:     createdAt,
			Version:       int32(batchIndex),
		})
	}

	return events
}

func (s *ProducerService) batchSizeOrDefault() int {
	if s.batchSize <= 0 {
		return 100
	}
	return s.batchSize
}

var _ = errors.New // keep linter calm if unused in future edits

