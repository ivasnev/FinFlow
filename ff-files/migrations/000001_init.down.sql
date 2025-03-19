-- Удаление индексов
DROP INDEX IF EXISTS idx_files_bucket;
DROP INDEX IF EXISTS idx_files_object_key;

-- Удаление таблицы
DROP TABLE IF EXISTS files; 