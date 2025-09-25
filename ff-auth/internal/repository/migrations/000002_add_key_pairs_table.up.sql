-- Таблица для хранения пар ключей ED25519
CREATE TABLE IF NOT EXISTS key_pairs (
    id SERIAL PRIMARY KEY,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создаем индекс для быстрого поиска активных ключей
CREATE INDEX IF NOT EXISTS idx_key_pairs_is_active ON key_pairs(is_active); 