-- Удаление индексов
DROP INDEX IF EXISTS idx_service_access_from_id;
DROP INDEX IF EXISTS idx_service_access_to_id;
DROP INDEX IF EXISTS idx_key_pairs_service_id;

-- Удаление таблиц
DROP TABLE IF EXISTS key_pairs;
DROP TABLE IF EXISTS service_access;
DROP TABLE IF EXISTS services; 