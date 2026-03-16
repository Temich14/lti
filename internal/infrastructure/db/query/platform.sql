-- name: CreatePlatform :one
INSERT INTO lti_platforms (issuer,
                           client_id,
                           deployment_id,
                           auth_endpoint,
                           token_endpoint,
                           jwks_uri,
                           registration_endpoint)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetPlatformByIssuerAndClientID :one
SELECT *
FROM lti_platforms
WHERE issuer = $1
  AND client_id = $2;

-- name: GetPlatformByID :one
SELECT *
FROM lti_platforms
WHERE id = $1;

-- name: ListPlatforms :many
SELECT *
FROM lti_platforms
ORDER BY created_at DESC;

-- name: DeletePlatform :exec
DELETE
FROM lti_platforms
WHERE id = $1;

-- name: UpdateClientID :exec
UPDATE lti_platforms
SET client_id = $2
WHERE issuer = $1;

ALTER TABLE lti_platforms
    ADD CONSTRAINT unique_platform UNIQUE (issuer, client_id);