-- Удаление индексов
DROP INDEX IF EXISTS idx_tasks_user_id;
DROP INDEX IF EXISTS idx_tasks_event_id;
DROP INDEX IF EXISTS idx_user_transaction_transaction_id;
DROP INDEX IF EXISTS idx_transactions_event_id;
DROP INDEX IF EXISTS idx_activities_id_event;
DROP INDEX IF EXISTS idx_events_category_id;

-- Удаление таблиц в обратном порядке создания
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS user_transaction;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS transaction_types;
DROP TABLE IF EXISTS icons;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS user_event;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users; 