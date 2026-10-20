CREATE TABLE settlements (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    from_user_id BIGINT NOT NULL REFERENCES users(id),
    to_user_id BIGINT NOT NULL REFERENCES users(id),
    created_by BIGINT NOT NULL REFERENCES users(id),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0 AND amount_cents <= 9999999999),
    request_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    canceled_at TIMESTAMPTZ,
    CHECK (from_user_id <> to_user_id),
    CHECK (created_by = to_user_id),
    UNIQUE(event_id, created_by, request_id)
);
CREATE INDEX idx_settlements_event ON settlements(event_id, id DESC);
CREATE INDEX idx_settlements_active ON settlements(event_id) WHERE canceled_at IS NULL;
