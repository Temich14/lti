package service

import (
	"context"

	"github.com/google/uuid"
)

// EnrollmentSyncStarter enqueues NRPS-backed roster synchronization for an LTI launch.
type EnrollmentSyncStarter interface {
	StartOrResumeRosterSync(ctx context.Context, issuer, clientID, lmsCourseID, nrpsContextMembershipsURL string) (uuid.UUID, error)
}
