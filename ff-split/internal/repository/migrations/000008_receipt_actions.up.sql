ALTER TABLE receipts ADD COLUMN deleted_at timestamptz;
ALTER TABLE receipts ADD COLUMN image_available boolean NOT NULL DEFAULT true;
UPDATE receipts SET image_available=false WHERE image_key IN (SELECT image_key FROM receipt_image_gc);
ALTER TABLE receipts ADD COLUMN queued_by bigint;
ALTER TABLE receipts ADD COLUMN queued_at timestamptz;
ALTER TABLE receipts ADD COLUMN retry_revision integer NOT NULL DEFAULT 0;
UPDATE receipts SET queued_by=created_by,queued_at=created_at;
ALTER TABLE receipts ALTER COLUMN queued_by SET NOT NULL;
ALTER TABLE receipts ALTER COLUMN queued_at SET NOT NULL;
ALTER TABLE receipts ALTER COLUMN queued_at SET DEFAULT clock_timestamp();
ALTER TABLE receipts DROP CONSTRAINT receipts_event_id_image_sha256_key;
CREATE UNIQUE INDEX receipts_active_hash ON receipts(event_id,image_sha256) WHERE deleted_at IS NULL;
DROP INDEX receipts_pending;
CREATE INDEX receipts_pending ON receipts(queued_by,event_id) WHERE deleted_at IS NULL AND status IN ('uploading','queued','processing');
CREATE INDEX receipts_activity ON receipts(updated_at) WHERE status!='confirmed';
CREATE TABLE receipt_admissions (
    id bigserial PRIMARY KEY,
    receipt_id uuid NOT NULL,
    actor_id bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX receipt_admissions_actor ON receipt_admissions(actor_id,created_at DESC);
CREATE INDEX receipt_admissions_age ON receipt_admissions(created_at);
INSERT INTO receipt_admissions(receipt_id,actor_id,created_at) SELECT id,created_by,created_at FROM receipts WHERE created_at>clock_timestamp()-interval '7 days';
-- Журнал допуска переживает удаление чека и события, чтобы очистка не обнуляла квоту.
CREATE FUNCTION receipt_admit_upload() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.queued_by IS NULL THEN NEW.queued_by:=NEW.created_by; END IF;
  INSERT INTO receipt_admissions(receipt_id,actor_id,created_at) VALUES(NEW.id,NEW.created_by,NEW.created_at);
  RETURN NEW;
END; $$;
CREATE TRIGGER receipt_admit_upload BEFORE INSERT ON receipts FOR EACH ROW EXECUTE FUNCTION receipt_admit_upload();
