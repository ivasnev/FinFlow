CREATE TABLE event_invitations (
 id BIGSERIAL PRIMARY KEY,
 event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
 token_hash CHAR(64) NOT NULL UNIQUE,
 created_by BIGINT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days'),
 revoked_at TIMESTAMPTZ
);
CREATE INDEX event_invitations_active ON event_invitations(event_id,created_by,expires_at) WHERE revoked_at IS NULL;
