-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS teams (
    id      SERIAL PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
    id           TEXT PRIMARY KEY,
    username     TEXT NOT NULL,
    team_id      INT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    is_active    BOOLEAN NOT NULL
);

CREATE TYPE pull_request_status AS ENUM ('OPEN', 'MERGED');
CREATE TABLE IF NOT EXISTS pull_requests (
    id                  TEXT PRIMARY KEY,
    pull_requests_name  TEXT NOT NULL,
    author_id           TEXT NOT NULL REFERENCES users(id),
    status              pull_request_status NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at           TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id     TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    reviewer_id         TEXT NOT NULL REFERENCES users(id),
    PRIMARY KEY (pull_request_id, reviewer_id)
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

DROP TYPE IF EXISTS pull_request_status;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS pull_request_reviewers;