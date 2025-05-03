-- Удаление таблиц в обратном порядке
drop table if exists debts cascade;
drop table if exists transaction_shares cascade;
drop table if exists transactions cascade;
drop table if exists transaction_categories cascade;
drop table if exists icons cascade;
drop table if exists activities cascade;
drop table if exists user_event cascade;
drop table if exists events cascade;
drop table if exists categories cascade;
drop table if exists users cascade;

-- Удаление индексов, если они существуют
drop index if exists idx_debts_transaction_id;
drop index if exists idx_transaction_shares_tx_id;
drop index if exists idx_transactions_event_id;
drop index if exists idx_activities_id_event;
drop index if exists idx_events_category_id;
drop index if exists idx_tasks_event_id;
drop index if exists idx_tasks_user_id;
