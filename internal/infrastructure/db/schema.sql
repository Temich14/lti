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