ALTER TABLE events ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX events_archive ON events(deleted_at,id) WHERE deleted_at IS NOT NULL;
CREATE FUNCTION event_purge_at(archived TIMESTAMPTZ) RETURNS TIMESTAMPTZ
 LANGUAGE SQL IMMUTABLE STRICT PARALLEL SAFE AS $$
 SELECT ((archived AT TIME ZONE 'Europe/Moscow') + INTERVAL '6 months') AT TIME ZONE 'Europe/Moscow';
$$;
