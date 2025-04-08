-- Создание базовой схемы для сервиса авторизации

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,
    nickname TEXT NOT NULL UNIQUE,
    name TEXT,
    birthdate DATE,
    avatar UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица аватарок пользователей
CREATE TABLE IF NOT EXISTS user_avatars (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_id UUID NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создаем внешний ключ от users.avatar к user_avatars.id
ALTER TABLE users
    ADD CONSTRAINT fk_users_avatar
    FOREIGN KEY (avatar) 
    REFERENCES user_avatars(id) 
    ON DELETE SET NULL;
