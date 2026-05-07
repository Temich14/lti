package enrollmentsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RosterSyncRun struct {
	RunID        uuid.UUID
	Issuer       string
	ClientID     string
	CourseID     string
	MembershipsURL string

	NRPSCursor       string
	NextBatchIndex  int32
	Version          int32
	SyncAttempts    int32

	LockedBy    string
	LockedUntil time.Time
}

type OutboxEvent struct {
	ID            uuid.UUID
	EventID      uuid.UUID
	PartitionKey string
	Envelope     json.RawMessage
	Attempts     int32
}

type ConsumerInboxKey struct {
	EventID uuid.UUID
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) StartOrResumeRosterSync(ctx context.Context, issuer, clientID, courseID, membershipsURL string) (uuid.UUID, error) {
	runID := uuid.New()

	// Start a new run, but if an active run already exists for (issuer, courseID), reuse it.
	// We rely on the partial unique index and handle the 23505 conflict.
	_, err := r.pool.Exec(ctx, `
		INSERT INTO roster_sync_runs (run_id, issuer, client_id, lms_course_id, nrps_context_memberships_url, nrps_cursor, next_batch_index, status, version)
		VALUES ($1, $2, $3, $4, $5, '', 0, 'running', 0)
	`, runID, issuer, clientID, courseID, membershipsURL)
	if err == nil {
		return runID, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		// Active run exists: return it, and optionally refresh memberships URL if changed.
		var existingID uuid.UUID
		if getErr := r.pool.QueryRow(ctx, `
			SELECT run_id FROM roster_sync_runs
			WHERE issuer = $1 AND lms_course_id = $2 AND status = 'running'
			ORDER BY updated_at DESC
			LIMIT 1
		`, issuer, courseID).Scan(&existingID); getErr != nil {
			return uuid.Nil, getErr
		}

		// Best-effort update; deterministic event_id requires run_id stability, not URL stability.
		_, _ = r.pool.Exec(ctx, `
			UPDATE roster_sync_runs
			SET nrps_context_memberships_url = $1, updated_at = now()
			WHERE run_id = $2
		`, membershipsURL, existingID)

		return existingID, nil
	}

	return uuid.Nil, fmt.Errorf("start roster sync: %w", err)
}

func (r *Repository) ClaimRosterSyncRun(ctx context.Context, owner string, lease time.Duration) (*RosterSyncRun, bool, error) {
	seconds := int64(lease.Seconds())
	var run RosterSyncRun

	// Lease-based claiming to allow multiple workers without long DB locks during NRPS HTTP calls.
	// If the worker crashes, another worker can re-claim after locked_until.
	err := r.pool.QueryRow(ctx, `
		UPDATE roster_sync_runs
		SET locked_by = $1,
		    locked_until = now() + ($2 * interval '1 second'),
		    updated_at = now()
		WHERE run_id = (
			SELECT run_id
			FROM roster_sync_runs
			WHERE status = 'running'
			  AND (locked_until IS NULL OR locked_until < now())
			ORDER BY updated_at ASC
			LIMIT 1
		)
		RETURNING run_id, issuer, client_id, lms_course_id,
		          nrps_context_memberships_url,
		          nrps_cursor, next_batch_index, version, sync_attempts,
		          locked_by, locked_until
	`, owner, seconds).Scan(
		&run.RunID, &run.Issuer, &run.ClientID, &run.CourseID,
		&run.MembershipsURL,
		&run.NRPSCursor, &run.NextBatchIndex, &run.Version, &run.SyncAttempts,
		&run.LockedBy, &run.LockedUntil,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &run, true, nil
}

// ---- Outbox: transactional inserts from sync worker ----

func (r *Repository) AppendOutboxEventsAndAdvanceRun(
	ctx context.Context,
	owner string,
	runID uuid.UUID,
	expectedVersion int32,
	nextCursor string,
	nextBatchIndex int32,
	newVersion int32,
	completed bool,
	events []IncomingEvent,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1) Persist outbox events for each batch.
	for _, ev := range events {
		envelopeBytes, err := json.Marshal(ev)
		if err != nil {
			return err
		}

		partitionKey := ev.AggregateID.String() // keep all batches for the same run in one Kafka partition

		_, err = tx.Exec(ctx, `
			INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, partition_key, envelope)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (event_id) DO NOTHING
		`, ev.ID, ev.AggregateType, ev.AggregateID, ev.EventType, partitionKey, envelopeBytes)
		if err != nil {
			return err
		}
	}

	// 2) Advance run progress and release the lease.
	tag, err := tx.Exec(ctx, `
		UPDATE roster_sync_runs
		SET nrps_cursor = $1,
		    next_batch_index = $2,
		    version = $3,
		    sync_attempts = 0,
		    last_error = NULL,
		    status = CASE WHEN $7 THEN 'completed' ELSE 'running' END,
		    completed_at = CASE WHEN $7 THEN now() ELSE NULL END,
		    locked_by = NULL,
		    locked_until = NULL,
		    updated_at = now()
		WHERE run_id = $4 AND version = $5 AND locked_by = $6
	`, nextCursor, nextBatchIndex, newVersion, runID, expectedVersion, owner, completed)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("roster sync run version conflict or lease stolen")
	}

	return tx.Commit(ctx)
}

func (r *Repository) MarkRunFailed(ctx context.Context, runID uuid.UUID, owner string, expectedVersion int32, errMsg string, backoff time.Duration, maxAttempts int32) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	newAttemptsQuery := `
		SELECT sync_attempts FROM roster_sync_runs WHERE run_id = $1 AND version = $2 AND locked_by = $3
	`
	var attempts int32
	if err := tx.QueryRow(ctx, newAttemptsQuery, runID, expectedVersion, owner).Scan(&attempts); err != nil {
		return err
	}
	attempts++

	newStatus := "running"
	if attempts >= maxAttempts {
		newStatus = "failed"
	}

	seconds := int64(backoff.Seconds())
	_, err = tx.Exec(ctx, `
		UPDATE roster_sync_runs
		SET sync_attempts = $1,
		    last_error = $2,
		    status = $3,
		    locked_until = now() + ($4 * interval '1 second'),
		    locked_by = $5,
		    updated_at = now()
		WHERE run_id = $6 AND version = $7 AND locked_by = $8
	`, attempts, errMsg, newStatus, seconds, owner, runID, expectedVersion, owner)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ---- Outbox: publication ----

func (r *Repository) ClaimOutboxEvents(ctx context.Context, owner string, limit int, lease time.Duration) ([]OutboxEvent, error) {
	seconds := int64(lease.Seconds())
	rows, err := r.pool.Query(ctx, `
		WITH cte AS (
			SELECT id
			FROM outbox_events
			WHERE processed_at IS NULL
			  AND (locked_until IS NULL OR locked_until < now())
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox_events o
		SET locked_by = $2,
		    locked_until = now() + ($3 * interval '1 second')
		FROM cte
		WHERE o.id = cte.id
		RETURNING o.id, o.event_id, o.partition_key, o.envelope, o.attempts
	`, limit, owner, seconds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventID, &e.PartitionKey, &e.Envelope, &e.Attempts); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) MarkOutboxProcessed(ctx context.Context, owner string, outboxID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events
		SET processed_at = now(),
		    locked_by = NULL,
		    locked_until = NULL,
		    last_error = NULL
		WHERE id = $1 AND locked_by = $2 AND processed_at IS NULL
	`, outboxID, owner)
	return err
}

func (r *Repository) MarkOutboxFailed(ctx context.Context, owner string, outboxID uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events
		SET attempts = attempts + 1,
		    last_error = $1,
		    locked_by = NULL,
		    locked_until = NULL
		WHERE id = $2 AND locked_by = $3 AND processed_at IS NULL
	`, errMsg, outboxID, owner)
	return err
}

func (r *Repository) MarkOutboxDeadLetter(ctx context.Context, owner string, outboxID uuid.UUID, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events
		SET processed_at = now(),
		    locked_by = NULL,
		    locked_until = NULL,
		    last_error = $1
		WHERE id = $2 AND locked_by = $3 AND processed_at IS NULL
	`, errMsg, outboxID, owner)
	return err
}

// ApplyUserSynchronizationEvent applies users synchronization batch with exactly-once effect.
// It uses the inbox table as a deduplication mechanism by event_id.
func (r *Repository) ApplyUserSynchronizationEvent(ctx context.Context, ev IncomingEvent) error {
	var payload UserSynchronizationPayloadV1
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1) Insert into inbox; if already exists, skip all side effects.
	res, err := tx.Exec(ctx, `
		INSERT INTO inbox_events (event_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_id) DO NOTHING
	`, ev.ID, ev.AggregateType, ev.AggregateID, ev.EventType, ev.Payload)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}

	issuer := payload.LMS.Issuer
	courseID := payload.LMS.CourseID

	// 2) Upsert users and course enrollments for each user in the batch.
	for _, u := range payload.Users {
		if strings.TrimSpace(u.Sub) == "" || strings.TrimSpace(u.Email) == "" {
			continue
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO lms_users (issuer, sub, email, display_name, avatar_url, version)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (issuer, sub) DO UPDATE
			SET version = CASE
			    	WHEN EXCLUDED.version > lms_users.version THEN EXCLUDED.version
			    	ELSE lms_users.version
			    END,
			    email = CASE
			    	WHEN EXCLUDED.version > lms_users.version THEN EXCLUDED.email
			    	ELSE lms_users.email
			    END,
			    display_name = CASE
			    	WHEN EXCLUDED.version > lms_users.version THEN EXCLUDED.display_name
			    	ELSE lms_users.display_name
			    END,
			    avatar_url = CASE
			    	WHEN EXCLUDED.version > lms_users.version THEN EXCLUDED.avatar_url
			    	ELSE lms_users.avatar_url
			    END,
			    updated_at = now()
		`, issuer, u.Sub, u.Email, u.DisplayName, u.AvatarURL, ev.Version)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO lms_course_enrollments (
				issuer, course_id, user_sub, user_email, display_name, avatar_url, enrolled, last_applied_version
			)
			VALUES ($1,$2,$3,$4,$5,$6,true,$7)
			ON CONFLICT (issuer, course_id, user_sub) DO UPDATE
			SET user_email = CASE
				WHEN EXCLUDED.last_applied_version > lms_course_enrollments.last_applied_version THEN EXCLUDED.user_email
				ELSE lms_course_enrollments.user_email
			END,
			display_name = CASE
				WHEN EXCLUDED.last_applied_version > lms_course_enrollments.last_applied_version THEN EXCLUDED.display_name
				ELSE lms_course_enrollments.display_name
			END,
			avatar_url = CASE
				WHEN EXCLUDED.last_applied_version > lms_course_enrollments.last_applied_version THEN EXCLUDED.avatar_url
				ELSE lms_course_enrollments.avatar_url
			END,
			last_applied_version = GREATEST(lms_course_enrollments.last_applied_version, EXCLUDED.last_applied_version),
			updated_at = now(),
			enrolled = true
		`, issuer, courseID, u.Sub, u.Email, u.DisplayName, u.AvatarURL, ev.Version)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE inbox_events
		SET processed_at = now(), last_error = NULL
		WHERE event_id = $1
	`, ev.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

