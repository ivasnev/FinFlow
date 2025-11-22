-- Миграции для тестовой базы данных

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	nickname TEXT NOT NULL UNIQUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица ролей
CREATE TABLE IF NOT EXISTS roles (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);

-- Таблица связи пользователей и ролей
CREATE TABLE IF NOT EXISTS user_roles (
	user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
	role_id INT REFERENCES roles(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, role_id)
);

-- Таблица сессий пользователей
CREATE TABLE IF NOT EXISTS sessions (
	id UUID PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	refresh_token TEXT NOT NULL UNIQUE,
	ip_address TEXT[],
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица истории входов
CREATE TABLE IF NOT EXISTS login_history (
	id SERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	ip_address INET NOT NULL,
	user_agent TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица устройств пользователей
CREATE TABLE IF NOT EXISTS devices (
	id SERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	device_id TEXT NOT NULL UNIQUE,
	user_agent TEXT NOT NULL,
	last_login TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица ключевых пар для токенов
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

-- Заполнение таблицы ролей начальными данными
INSERT INTO roles (name) VALUES 
	('admin'),
	('user'),
	('moderator')
ON CONFLICT (name) DO NOTHING;

