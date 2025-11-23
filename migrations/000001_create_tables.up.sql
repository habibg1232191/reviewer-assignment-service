CREATE TABLE team (
    team_name TEXT PRIMARY KEY
);

CREATE TABLE users (
    user_id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    team_name TEXT NOT NULL REFERENCES team(team_name),
    is_active BOOLEAN NOT NULL
);

CREATE TABLE pull_request (
    pr_id TEXT PRIMARY KEY,
    pr_name TEXT NOT NULL UNIQUE,
    author_id TEXT NOT NULL REFERENCES users(user_id),
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pr_reviewers (
    pr_id TEXT NOT NULL REFERENCES pull_request(pr_id) ON DELETE CASCADE,
    reviewer_id TEXT NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (pr_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);

CREATE INDEX idx_user_id ON users(user_id);
CREATE INDEX idx_is_active ON users(is_active);

CREATE INDEX idx_pr_status ON pull_request(status);
CREATE INDEX idx_pr_author_id ON pull_request(author_id);
