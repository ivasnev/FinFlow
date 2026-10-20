DROP FUNCTION event_purge_at(TIMESTAMPTZ);
DROP INDEX events_archive;
ALTER TABLE events DROP COLUMN deleted_at;
