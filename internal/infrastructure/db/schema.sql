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