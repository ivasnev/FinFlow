CREATE TABLE receipts (
    sequence bigserial UNIQUE NOT NULL,
    id uuid PRIMARY KEY,
    event_id bigint NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_by bigint NOT NULL,
    image_key text NOT NULL UNIQUE,
    image_sha256 char(64) NOT NULL,
    content_type text NOT NULL,
    status text NOT NULL DEFAULT 'queued' CHECK(status IN ('uploading','queued','processing','ready','failed','confirmed')),
    revision integer NOT NULL DEFAULT 1,
    attempts integer NOT NULL DEFAULT 0,
    lease_token uuid,
    lease_until timestamptz,
    available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    result jsonb NOT NULL DEFAULT '{}',
    draft jsonb NOT NULL DEFAULT '{}',
    transaction_ids jsonb NOT NULL DEFAULT '[]',
    error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE(event_id,image_sha256)
);
CREATE INDEX receipts_queue ON receipts(available_at,sequence) WHERE status IN ('queued','processing');
CREATE INDEX receipts_event ON receipts(event_id,sequence DESC);
CREATE INDEX receipts_quota ON receipts(created_by,created_at DESC);
CREATE INDEX receipts_pending ON receipts(created_by,event_id) WHERE status IN ('uploading','queued','processing');
CREATE INDEX receipts_created ON receipts(created_at);
CREATE TABLE receipt_image_gc (image_key text PRIMARY KEY, created_at timestamptz NOT NULL DEFAULT clock_timestamp());
CREATE FUNCTION receipt_image_deleted() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO receipt_image_gc(image_key) VALUES(OLD.image_key) ON CONFLICT DO NOTHING;
  RETURN OLD;
END; $$;
CREATE TRIGGER receipt_image_deleted BEFORE DELETE ON receipts FOR EACH ROW EXECUTE FUNCTION receipt_image_deleted();
