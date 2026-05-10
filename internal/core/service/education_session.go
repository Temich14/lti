package service

import "context"

// EducationSessionCreateParams is the cross-service contract for registering a learning session on LTI launch.
type EducationSessionCreateParams struct {
	RequestID  string
	UserID     string
	CourseID   string
	ResourceID string
}

// EducationSessionCreator opens an educational session via an external backend (typically gRPC).
type EducationSessionCreator interface {
	CreateForLTILaunch(ctx context.Context, p EducationSessionCreateParams) (remoteSessionID string, err error)
}
