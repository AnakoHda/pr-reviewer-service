-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_users_by_team_id ON users(team_id);

CREATE INDEX IF NOT EXISTS idx_pull_requests_by_author_id ON pull_requests(author_id);

CREATE INDEX IF NOT EXISTS idx_pull_request_reviewers_by_reviewer_id ON pull_request_reviewers(reviewer_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX IF EXISTS idx_users_by_team_id;
DROP INDEX IF EXISTS idx_pull_requests_by_author_id;
DROP INDEX IF EXISTS idx_pull_request_reviewers_by_reviewer_id;