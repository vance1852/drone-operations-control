CREATE TABLE console_sessions (
    token uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES console_users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX console_sessions_user_expiry_idx ON console_sessions(user_id, expires_at)
    WHERE revoked_at IS NULL;

INSERT INTO console_users(id, username, password_hash, real_name, phone, role, status)
VALUES ('10000000-0000-0000-0000-000000000002', 'operator',
        'ec6e1c25258002eb1c67d15c7f45da7945fa4c58778fd7d88faa5e53e3b4698d',
        '飞行调度员', '13800138001', 0, 1)
ON CONFLICT (username) DO NOTHING;
