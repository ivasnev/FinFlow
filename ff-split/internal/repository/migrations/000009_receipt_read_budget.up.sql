-- Private image reads consume a durable global budget before any S3 request.
CREATE TABLE receipt_image_reads (
 id BIGSERIAL PRIMARY KEY,
 actor_id BIGINT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX receipt_image_reads_age ON receipt_image_reads(created_at);
CREATE INDEX receipt_image_reads_actor ON receipt_image_reads(actor_id,created_at DESC);
