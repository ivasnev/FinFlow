-- Миграции для тестовой базы данных ff-id

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

-- Таблица связей между пользователями (друзья)
CREATE TABLE IF NOT EXISTS user_friends (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Добавляем уникальный индекс, чтобы предотвратить дублирование связей
    CONSTRAINT unique_user_friend UNIQUE (user_id, friend_id),
    
    -- Проверка, чтобы пользователь не мог добавить сам себя в друзья
    CONSTRAINT user_not_friend_self CHECK (user_id <> friend_id)
);

-- Создаем индексы для быстрого поиска друзей
CREATE INDEX IF NOT EXISTS idx_user_friends_user_id ON user_friends(user_id);
CREATE INDEX IF NOT EXISTS idx_user_friends_friend_id ON user_friends(friend_id);
CREATE INDEX IF NOT EXISTS idx_user_friends_status ON user_friends(status);

