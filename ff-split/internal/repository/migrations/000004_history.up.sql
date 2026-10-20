CREATE TABLE event_history (
 id BIGSERIAL PRIMARY KEY,
 event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
 actor_id BIGINT NOT NULL,
 actor_name TEXT NOT NULL,
 kind TEXT NOT NULL,
 entity_id BIGINT NOT NULL,
 before_values JSONB NOT NULL,
 after_values JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX event_history_page ON event_history(event_id,id DESC);
