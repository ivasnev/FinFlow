-- Сохраняем существующую UTC-интерпретацию старых значений.
-- Новые сроки хранятся как абсолютное время независимо от часового пояса сервиса.
ALTER TABLE sessions ALTER COLUMN expires_at TYPE TIMESTAMPTZ
    USING expires_at AT TIME ZONE 'UTC';
