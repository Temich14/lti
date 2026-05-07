CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE lti_platforms
(
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer                text NOT NULL,
    client_id             text NOT NULL,
    deployment_id         text,
    auth_endpoint         text,
    token_endpoint        text,
    jwks_uri              text,
    registration_endpoint text,
    created_at            timestamptz      DEFAULT now()
);

create table courses
(
    id uuid primary key,
    name text not null,
    icon_url text,
    version int not null
);
create table modules
(
    id uuid primary key,
    name text not null,
    course id not null,
    version int not null
);
create table lessons
(
    id uuid primary key,
    name text not null,
    module id not null,
    version int not null
);

create table inbox
(
    id uuid primary key
);

-- Enrollment synchronization (NRPS roster -> Kafka -> consumer inbox)
CREATE TABLE IF NOT EXISTS roster_sync_runs
(
    run_id                         uuid PRIMARY KEY,
    issuer                         text      NOT NULL,
    client_id                      text      NOT NULL,
    lms_course_id                  text      NOT NULL,
    nrps_context_memberships_url  text      NOT NULL,
    nrps_cursor                    text      NOT NULL DEFAULT '',
    next_batch_index              int4      NOT NULL DEFAULT 0,
    status                         text      NOT NULL DEFAULT 'running', -- running|completed|failed
    last_error                     text,
    sync_attempts                 int4      NOT NULL DEFAULT 0,
    version                        int4      NOT NULL DEFAULT 0,
    locked_by                      text,
    locked_until                   timestamptz,
    created_at                     timestamptz NOT NULL DEFAULT now(),
    updated_at                     timestamptz NOT NULL DEFAULT now(),
    completed_at                   timestamptz
);

-- Ensure a single active run per (issuer, course) to keep event_id deterministic.
CREATE UNIQUE INDEX IF NOT EXISTS roster_sync_runs_active_unique
    ON roster_sync_runs (issuer, lms_course_id)
    WHERE status = 'running';

CREATE INDEX IF NOT EXISTS roster_sync_runs_running_idx ON roster_sync_runs(status, updated_at);

CREATE TABLE IF NOT EXISTS outbox_events
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id       uuid        NOT NULL UNIQUE,
    aggregate_type text        NOT NULL,
    aggregate_id   uuid        NOT NULL,
    event_type      text       NOT NULL,
    partition_key  text       NOT NULL,
    envelope       jsonb      NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    processed_at   timestamptz,
    attempts       int4        NOT NULL DEFAULT 0,
    last_error     text,
    locked_by      text,
    locked_until   timestamptz
);

CREATE INDEX IF NOT EXISTS outbox_pending_idx
    ON outbox_events(processed_at, locked_until, created_at);

-- Consumer inbox pattern: deduplicate by event_id to achieve exactly-once effect.
CREATE TABLE IF NOT EXISTS inbox_events
(
    event_id        uuid PRIMARY KEY,
    aggregate_type  text NOT NULL,
    aggregate_id    uuid NOT NULL,
    event_type      text NOT NULL,
    payload         jsonb NOT NULL,
    received_at     timestamptz NOT NULL DEFAULT now(),
    processed_at    timestamptz,
    last_error      text
);

-- Minimal persistence for applied enrollment synchronization.
CREATE TABLE IF NOT EXISTS lms_users
(
    user_id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer        text NOT NULL,
    sub            text NOT NULL,
    email          text NOT NULL,
    display_name  text,
    avatar_url    text,
    version        int4 NOT NULL DEFAULT 0,
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (issuer, sub)
);

CREATE INDEX IF NOT EXISTS lms_users_issuer_email_idx ON lms_users(issuer, email);

CREATE TABLE IF NOT EXISTS lms_course_enrollments
(
    enrollment_id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer                text NOT NULL,
    course_id            text NOT NULL,
    user_sub             text NOT NULL,
    user_email          text NOT NULL,
    display_name        text,
    avatar_url          text,
    enrolled             boolean NOT NULL DEFAULT true,
    last_applied_version int4 NOT NULL DEFAULT -1,
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (issuer, course_id, user_sub)
);

CREATE INDEX IF NOT EXISTS lms_course_enrollments_course_idx ON lms_course_enrollments(issuer, course_id);

-- AGS (LTI Advantage Assignment & Grade Services) persistence
CREATE TABLE IF NOT EXISTS ags_line_items
(
    lineitem_id      uuid PRIMARY KEY,
    context_id       text NOT NULL,
    label            text NOT NULL,
    max_score        numeric(12,4) NOT NULL,
    start_date_time  timestamptz,
    end_date_time    timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ags_line_items_context_idx ON ags_line_items(context_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ags_scores
(
    lineitem_id      uuid NOT NULL REFERENCES ags_line_items(lineitem_id) ON DELETE CASCADE,
    user_id          text NOT NULL,
    score            numeric(12,4) NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (lineitem_id, user_id)
);

CREATE INDEX IF NOT EXISTS ags_scores_user_idx ON ags_scores(user_id, updated_at DESC);