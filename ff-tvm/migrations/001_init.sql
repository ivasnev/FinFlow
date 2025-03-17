-- Создание таблицы сервисов
CREATE TABLE IF NOT EXISTS services (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    private_key_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы доступа между сервисами
CREATE TABLE IF NOT EXISTS service_access (
    id BIGSERIAL PRIMARY KEY,
    from_id BIGINT NOT NULL REFERENCES services(id),
    to_id BIGINT NOT NULL REFERENCES services(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_id, to_id)
);

-- Создание таблицы пар ключей
CREATE TABLE IF NOT EXISTS key_pairs (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_service_access_from_id ON service_access(from_id);
CREATE INDEX IF NOT EXISTS idx_service_access_to_id ON service_access(to_id);
CREATE INDEX IF NOT EXISTS idx_key_pairs_service_id ON key_pairs(service_id); 