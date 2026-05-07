package enrollmentsync

import (
	"encoding/json"
	"fmt"
	"time"

	core "LTICore/internal/core/domain"

	"github.com/google/uuid"
)

// IncomingEvent is a platform-level envelope contract for synchronization events.
// It is intentionally JSON-friendly to make outbox + retries transparent.
type IncomingEvent struct {
	ID             uuid.UUID       `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
	Version       int32           `json:"version"`
}

type UserSynchronizationUser struct {
	Sub          string `json:"sub"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url"`
}

type UserSynchronizationPayloadV1 struct {
	LMS struct {
		Issuer    string `json:"issuer"`
		CourseID string `json:"course_id"`
	} `json:"lms"`
	Batch struct {
		Index int `json:"index"`
		Total int `json:"total,omitempty"`
	} `json:"batch"`
	Users []UserSynchronizationUser `json:"users"`
}

const (
	AggregateTypeRosterSyncRun           = "RosterSyncRun"
	EventTypeUserSynchronizationBatchV1   = "UserSynchronizationBatchV1"
	KafkaTopicUsersSynchronizationV1      = "users.synchronization.v1"
)

var eventIDNamespace = uuid.MustParse("6d3d8b2f-4d2c-4dd0-9e9c-0c4a5d8e7a12")

// BatchEventID is deterministic for a given sync run + batch index.
// This is the cornerstone for exactly-once effect via the consumer inbox.
func BatchEventID(runID uuid.UUID, batchIndex int32) uuid.UUID {
	name := fmt.Sprintf("users.synchronization.v1|%s|batch_index=%d", runID.String(), batchIndex)
	return uuid.NewSHA1(eventIDNamespace, []byte(name))
}

func MapNRPSMembersToUserProfiles(users []core.NRPSMember) []UserSynchronizationUser {
	out := make([]UserSynchronizationUser, 0, len(users))
	for _, m := range users {
		display := m.Name
		if display == "" && (m.GivenName != "" || m.FamilyName != "") {
			// Basic display name normalization for consumer.
			display = fmt.Sprintf("%s %s", m.GivenName, m.FamilyName)
		}
		out = append(out, UserSynchronizationUser{
			Sub:         m.UserID,
			Email:       m.Email,
			DisplayName: display,
			AvatarURL:   "",
		})
	}
	return out
}

