\set ON_ERROR_STOP on
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '60s';
SELECT pg_try_advisory_xact_lock(20581407,1) AS job_locked \gset
\if :job_locked
CREATE TEMP TABLE expired_events(id BIGINT PRIMARY KEY) ON COMMIT DROP;
INSERT INTO expired_events
 SELECT id FROM events
 WHERE deleted_at IS NOT NULL AND event_purge_at(deleted_at) <= clock_timestamp()
 ORDER BY deleted_at,id LIMIT 100 FOR UPDATE SKIP LOCKED;
DELETE FROM expired_events x WHERE NOT EXISTS (SELECT 1 FROM events e WHERE e.id=x.id AND e.deleted_at IS NOT NULL AND event_purge_at(e.deleted_at)<=clock_timestamp());
CREATE TEMP TABLE expired_guests(id BIGINT PRIMARY KEY) ON COMMIT DROP;
INSERT INTO expired_guests
 SELECT DISTINCT u.id FROM users u JOIN (
  SELECT user_id AS id FROM user_event WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT payer_id FROM transactions WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT s.user_id FROM transaction_shares s JOIN transactions t ON t.id=s.transaction_id WHERE t.event_id IN (SELECT id FROM expired_events)
  UNION SELECT d.from_user_id FROM debts d JOIN transactions t ON t.id=d.transaction_id WHERE t.event_id IN (SELECT id FROM expired_events)
  UNION SELECT d.to_user_id FROM debts d JOIN transactions t ON t.id=d.transaction_id WHERE t.event_id IN (SELECT id FROM expired_events)
  UNION SELECT from_user_id FROM optimized_debts WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT to_user_id FROM optimized_debts WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT user_id FROM activities WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT user_id FROM tasks WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT from_user_id FROM settlements WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT to_user_id FROM settlements WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT created_by FROM settlements WHERE event_id IN (SELECT id FROM expired_events)
  UNION SELECT entity_id FROM event_history WHERE event_id IN (SELECT id FROM expired_events) AND kind IN ('member.added','member.removed')
 ) referenced ON referenced.id=u.id WHERE u.is_dummy;
DELETE FROM debts WHERE transaction_id IN (SELECT id FROM transactions WHERE event_id IN (SELECT id FROM expired_events));
DELETE FROM transaction_shares WHERE transaction_id IN (SELECT id FROM transactions WHERE event_id IN (SELECT id FROM expired_events));
DELETE FROM transactions WHERE event_id IN (SELECT id FROM expired_events);
DELETE FROM optimized_debts WHERE event_id IN (SELECT id FROM expired_events);
DELETE FROM activities WHERE event_id IN (SELECT id FROM expired_events);
DELETE FROM tasks WHERE event_id IN (SELECT id FROM expired_events);
DELETE FROM user_event WHERE event_id IN (SELECT id FROM expired_events);
DELETE FROM events WHERE id IN (SELECT id FROM expired_events);
DELETE FROM users u WHERE u.is_dummy AND u.id IN (SELECT id FROM expired_guests)
 AND NOT EXISTS (SELECT 1 FROM user_event WHERE user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM transactions WHERE payer_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM transaction_shares WHERE user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM debts WHERE from_user_id=u.id OR to_user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM optimized_debts WHERE from_user_id=u.id OR to_user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM tasks WHERE user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM activities WHERE user_id=u.id)
 AND NOT EXISTS (SELECT 1 FROM settlements WHERE from_user_id=u.id OR to_user_id=u.id OR created_by=u.id);
SELECT count(*) FROM expired_events;
COMMIT;
\else
ROLLBACK;
SELECT -1;
\endif
